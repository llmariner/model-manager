package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-logr/logr"
	mv1 "github.com/llmariner/model-manager/api/v1"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/loader/internal/config"
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
	GetModelPath(ctx context.Context, in *mv1.GetModelPathRequest, opts ...grpc.CallOption) (*mv1.GetModelPathResponse, error)
	RegisterModel(ctx context.Context, in *mv1.RegisterModelRequest, opts ...grpc.CallOption) (*mv1.RegisterModelResponse, error)
	PublishModel(ctx context.Context, in *mv1.PublishModelRequest, opts ...grpc.CallOption) (*mv1.PublishModelResponse, error)
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

// GetModelPath gets the path of a model.
func (c *FakeModelClient) GetModelPath(ctx context.Context, in *mv1.GetModelPathRequest, opts ...grpc.CallOption) (*mv1.GetModelPathResponse, error) {
	path, ok := c.pathsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "model %q not found", in.Id)
	}
	return &mv1.GetModelPathResponse{
		Path: path,
	}, nil
}

// RegisterModel register a model.
func (c *FakeModelClient) RegisterModel(ctx context.Context, in *mv1.RegisterModelRequest, opts ...grpc.CallOption) (*mv1.RegisterModelResponse, error) {
	if _, ok := c.pathsByID[in.Id]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", in.Id)
	}
	// path := "models/default-tenant-id/" + in.Id
	c.pathsByID[in.Id] = in.Path
	return &mv1.RegisterModelResponse{
		Id:   in.Id,
		Path: in.Path,
	}, nil
}

// PublishModel publishes a model.
func (c *FakeModelClient) PublishModel(ctx context.Context, in *mv1.PublishModelRequest, opts ...grpc.CallOption) (*mv1.PublishModelResponse, error) {
	return &mv1.PublishModelResponse{}, nil
}

// New creates a new loader.
func New(
	baseModels []string,
	models []config.ModelConfig,
	objectStorePathPrefix string,
	baseModelPathPrefix string,
	modelDownloader ModelDownloader,
	s3Client S3Client,
	modelClient ModelClient,
	log logr.Logger,
) *L {
	return &L{
		baseModels:            baseModels,
		models:                models,
		objectStorePathPrefix: objectStorePathPrefix,
		baseModelPathPrefix:   baseModelPathPrefix,
		modelDownloader:       modelDownloader,
		s3Client:              s3Client,
		modelClient:           modelClient,
		log:                   log.WithName("loader"),
		tmpDir:                "/tmp",
	}
}

// L is a loader.
type L struct {
	baseModels []string

	models []config.ModelConfig

	// objectStorePathPrefix is the prefix of the path to the base and non-base models in the object stoer.
	objectStorePathPrefix string
	baseModelPathPrefix   string

	modelDownloader ModelDownloader

	s3Client S3Client

	modelClient ModelClient

	log logr.Logger

	tmpDir string
}

// Run runs the loader.
func (l *L) Run(ctx context.Context, interval time.Duration) error {
	if err := l.LoadModels(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := l.LoadModels(ctx); err != nil {
				return err
			}
		}
	}
}

// LoadModels loads base and non-base models.
func (l *L) LoadModels(ctx context.Context) error {
	for _, baseModel := range l.baseModels {
		if err := l.loadBaseModel(ctx, baseModel); err != nil {
			return err
		}
	}
	for _, m := range l.models {
		if err := l.loadModel(ctx, m); err != nil {
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
		l.log.Info("Already model exists", "modelID", convertedModelID)
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	mpath, modelInfo, err := l.downloadAndUploadModel(ctx, modelID, true)
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

	l.log.Info("Successfully loaded base model", "model", modelID)
	return nil
}

func (l *L) loadModel(ctx context.Context, model config.ModelConfig) error {
	if err := l.loadBaseModel(ctx, model.BaseModel); err != nil {
		return err
	}

	// HuggingFace uses '/" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	convertedModelID := strings.ReplaceAll(model.Model, "/", "-")
	convertedBaseModelID := strings.ReplaceAll(model.BaseModel, "/", "-")

	// First check if the model exists in the database.
	ctx = auth.AppendWorkerAuthorization(ctx)
	_, err := l.modelClient.GetModelPath(ctx, &mv1.GetModelPathRequest{Id: convertedModelID})
	if err == nil {
		l.log.Info("Already model exists", "modelID", convertedModelID)
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	mPath, _, err := l.downloadAndUploadModel(ctx, model.Model, false)
	if err != nil {
		return err
	}

	if _, err := l.modelClient.RegisterModel(ctx, &mv1.RegisterModelRequest{
		Id:           convertedModelID,
		BaseModel:    convertedBaseModelID,
		Adapter:      config.ToAdapterType(model.AdapterType),
		Quantization: config.ToQuantizationType(model.QuantizationType),
		Path:         mPath,
		// TODO(guangrui): Allow to configure project and org for models.
		ProjectId:      "default",
		OrganizationId: "default",
	}); err != nil {
		return err
	}
	if _, err := l.modelClient.PublishModel(ctx, &v1.PublishModelRequest{Id: convertedModelID}); err != nil {
		return err
	}
	l.log.Info("Successfully loaded, registered, and published model", "model", model.Model)
	return nil
}

type modelInfo struct {
	ggufModelPath            string
	hasHuggingFaceConfigJSON bool
}

func (l *L) downloadAndUploadModel(ctx context.Context, modelID string, isBaseModel bool) (string, *modelInfo, error) {
	log := l.log.WithValues("modelID", modelID)
	log.Info("Started loading model")

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
	log.Info("Created a temp dir", "path", tmpDir)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	log.Info("Downloading model")
	if err := l.modelDownloader.download(ctx, modelID, tmpDir); err != nil {
		return "", nil, err
	}

	toKey := func(path string) string {
		// Remove the tmpdir path from the path.
		relativePath := strings.TrimPrefix(path, tmpDir)
		return filepath.Join(l.toPathPrefix(isBaseModel), modelID, relativePath)
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

		if filepath.Base(path) == "config.json" || filepath.Base(path) == "adapter_config.json" {
			hasHuggingFaceConfigJSON = true
		}

		return nil
	}); err != nil {
		return "", nil, err
	}
	log.Info("Downloaded files", "count", len(paths))
	if len(paths) == 0 {
		return "", nil, fmt.Errorf("no files downloaded")
	}

	log.Info("Uploading model to the object store")
	for _, path := range paths {
		log.Info("Uploading", "path", path)
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

	mpath := filepath.Join(l.toPathPrefix(isBaseModel), modelID)
	return mpath, &modelInfo{
		ggufModelPath:            ggufModelPath,
		hasHuggingFaceConfigJSON: hasHuggingFaceConfigJSON,
	}, nil
}

func (l *L) toPathPrefix(isBaseModel bool) string {
	if isBaseModel {
		return filepath.Join(l.objectStorePathPrefix, l.baseModelPathPrefix)
	}
	// TODO(guangrui): make tenant-id configurable. The path should match with the path generated in RegisterModel.
	return filepath.Join(l.objectStorePathPrefix, "default-tenant-id")
}
