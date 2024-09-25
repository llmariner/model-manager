package loader

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mv1 "github.com/llm-operator/model-manager/api/v1"
	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ModelDownloader is an interface for downloading a model.
type ModelDownloader interface {
	download(ctx context.Context, modelName, destDir string) error
}

// NoopModelDownloader is a no-op model downloader.
type NoopModelDownloader struct {
}

func (d *NoopModelDownloader) download(ctx context.Context, modelName, destDir string) error {
	return nil
}

// S3Client is an interface for uploading a file to S3.
type S3Client interface {
	Upload(ctx context.Context, r io.Reader, key string) error
}

// NoopS3Client is a no-op S3 client.
type NoopS3Client struct {
}

// Upload uploads a file to S3.
func (c *NoopS3Client) Upload(ctx context.Context, r io.Reader, key string) error {
	return nil
}

// ModelClient is an interface for the model client.
type ModelClient interface {
	CreateBaseModel(ctx context.Context, in *mv1.CreateBaseModelRequest, opts ...grpc.CallOption) (*mv1.BaseModel, error)
	GetBaseModelPath(ctx context.Context, in *mv1.GetBaseModelPathRequest, opts ...grpc.CallOption) (*mv1.GetBaseModelPathResponse, error)
}

// NewFakeModelClient creates a fake model client.
func NewFakeModelClient() *FakeModelClient {
	return &FakeModelClient{
		pathsByID:   map[string]string{},
		ggufsByID:   map[string]string{},
		formatsByID: map[string][]mv1.ModelFormat{},
	}
}

// FakeModelClient is a fake model client.
type FakeModelClient struct {
	pathsByID   map[string]string
	ggufsByID   map[string]string
	formatsByID map[string][]mv1.ModelFormat
}

// CreateBaseModel creates a base model.
func (c *FakeModelClient) CreateBaseModel(ctx context.Context, in *mv1.CreateBaseModelRequest, opts ...grpc.CallOption) (*mv1.BaseModel, error) {
	if _, ok := c.pathsByID[in.Id]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", in.Id)
	}
	c.pathsByID[in.Id] = in.Path
	c.ggufsByID[in.Id] = in.GgufModelPath
	c.formatsByID[in.Id] = in.Formats

	return &mv1.BaseModel{
		Id: in.Id,
	}, nil
}

// GetBaseModelPath gets the path of a base model.
func (c *FakeModelClient) GetBaseModelPath(ctx context.Context, in *mv1.GetBaseModelPathRequest, opts ...grpc.CallOption) (*mv1.GetBaseModelPathResponse, error) {
	path, ok := c.pathsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "model %q not found", in.Id)
	}
	ggufPath, ok := c.ggufsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "GGUF model %q not found", in.Id)
	}
	formats, ok := c.formatsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "formats for model %q not found", in.Id)
	}

	return &mv1.GetBaseModelPathResponse{
		Path:          path,
		Formats:       formats,
		GgufModelPath: ggufPath,
	}, nil
}

// New creates a new loader.
func New(
	baseModels []string,
	objectStorePathPrefix string,
	modelDownloader ModelDownloader,
	s3Client S3Client,
	modelClient ModelClient,
) *L {
	return &L{
		baseModels:            baseModels,
		objectStorePathPrefix: objectStorePathPrefix,
		modelDownloader:       modelDownloader,
		s3Client:              s3Client,
		modelClient:           modelClient,
		tmpDir:                "/tmp",
	}
}

// L is a loader.
type L struct {
	baseModels []string

	// objectStorePathPrefix is the prefix of the path to the base models in the object stoer.
	objectStorePathPrefix string

	modelDownloader ModelDownloader

	s3Client S3Client

	modelClient ModelClient

	tmpDir string
}

// Run runs the loader.
func (l *L) Run(ctx context.Context, interval time.Duration) error {
	if err := l.LoadBaseModels(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := l.LoadBaseModels(ctx); err != nil {
				return err
			}
		}
	}
}

// LoadBaseModels loads the base models.
func (l *L) LoadBaseModels(ctx context.Context) error {
	for _, baseModel := range l.baseModels {
		if err := l.loadBaseModel(ctx, baseModel); err != nil {
			return err
		}
	}
	return nil
}

func (l *L) loadBaseModel(ctx context.Context, modelID string) error {
	// HuggingFace uses '/" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	convertedModelID := strings.ReplaceAll(modelID, "/", "-")

	// First check if the model exists in the database.
	ctx = auth.AppendWorkerAuthorization(ctx)
	_, err := l.modelClient.GetBaseModelPath(ctx, &mv1.GetBaseModelPathRequest{Id: convertedModelID})
	if err == nil {
		log.Printf("Model %q exists. Do nothing.\n", convertedModelID)
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	mpath, modelInfo, err := l.downloadAndUploadModel(ctx, modelID)
	if err != nil {
		return err
	}

	var formats []v1.ModelFormat
	if modelInfo.ggufModelPath != "" {
		formats = append(formats, mv1.ModelFormat_MODEL_FORMAT_GGUF)
	}
	if modelInfo.hasHuggingFaceConfigJSON {
		formats = append(formats, mv1.ModelFormat_MODEL_FORMAT_HUGGING_FACE)
	}

	if _, err := l.modelClient.CreateBaseModel(ctx, &mv1.CreateBaseModelRequest{
		Id:            convertedModelID,
		Path:          mpath,
		Formats:       formats,
		GgufModelPath: modelInfo.ggufModelPath,
	}); err != nil {
		return err
	}

	log.Printf("Successfully loaded base model %q\n", modelID)
	return nil
}

type modelInfo struct {
	ggufModelPath            string
	hasHuggingFaceConfigJSON bool
}

func (l *L) downloadAndUploadModel(ctx context.Context, modelID string) (string, *modelInfo, error) {
	log.Printf("Started loading base model %q\n", modelID)

	// Please note that the temp directory shouldn't contain a symlink. Otherwise
	// symlinks created by Hugging Face doesn't work.
	//
	// For example, suppose that
	// - /tmp is a symlink to private/tmp
	// - the temp dir /tmp/base-model0 is created.
	// - one of the symlinks reated by Hugging Face is .gitattributes, which is linked to ../../Users/kenji/.cache/.
	//
	// Then, the link does not work since /private/tmp/base-model0/../../Users/kenji/.cache/ is not a valid path.
	tmpDir, err := os.MkdirTemp(l.tmpDir, "base-model")
	if err != nil {
		return "", nil, err
	}
	log.Printf("Created a temp dir %q\n", tmpDir)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	log.Printf("Downloading base model %q\n", modelID)
	if err := l.modelDownloader.download(ctx, modelID, tmpDir); err != nil {
		return "", nil, err
	}

	toKey := func(path string) string {
		// Remove the tmpdir path from the path.
		relativePath := strings.TrimPrefix(path, tmpDir)
		return filepath.Join(l.objectStorePathPrefix, modelID, relativePath)
	}

	var (
		paths                    []string
		ggufModelPath            string
		hasHuggingFaceConfigJSON bool
	)
	if err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		paths = append(paths, path)

		if strings.HasSuffix(path, ".gguf") {
			if ggufModelPath != "" {
				return fmt.Errorf("multiple GGUF files found: %q and %q", ggufModelPath, path)
			}
			ggufModelPath = toKey(path)
		}

		if filepath.Base(path) == "config.json" {
			hasHuggingFaceConfigJSON = true
		}

		return nil
	}); err != nil {
		return "", nil, err
	}
	log.Printf("Downloaded %d files\n", len(paths))
	if len(paths) == 0 {
		return "", nil, fmt.Errorf("no files downloaded")
	}

	log.Printf("Uploading base model %q to the object store\n", modelID)
	for _, path := range paths {
		log.Printf("Uploading %q\n", path)
		r, err := os.Open(path)
		if err != nil {
			return "", nil, err
		}

		if err := l.s3Client.Upload(ctx, r, toKey(path)); err != nil {
			return "", nil, err
		}
		if err := r.Close(); err != nil {
			return "", nil, err
		}
	}

	mpath := filepath.Join(l.objectStorePathPrefix, modelID)
	return mpath, &modelInfo{
		ggufModelPath:            ggufModelPath,
		hasHuggingFaceConfigJSON: hasHuggingFaceConfigJSON,
	}, nil
}
