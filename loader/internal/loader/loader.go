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
	download(ctx context.Context, modelName, filename, destDir string) error
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

	CreateHFModelRepo(ctx context.Context, in *mv1.CreateHFModelRepoRequest, opts ...grpc.CallOption) (*mv1.HFModelRepo, error)
	GetHFModelRepo(ctx context.Context, in *mv1.GetHFModelRepoRequest, opts ...grpc.CallOption) (*mv1.HFModelRepo, error)
}

// NewFakeModelClient creates a fake model client.
func NewFakeModelClient() *FakeModelClient {
	return &FakeModelClient{
		pathsByID:   map[string]string{},
		ggufsByID:   map[string]string{},
		formatsByID: map[string][]mv1.ModelFormat{},

		hfModelRepos: map[string]bool{},
	}
}

// FakeModelClient is a fake model client.
type FakeModelClient struct {
	pathsByID   map[string]string
	ggufsByID   map[string]string
	formatsByID map[string][]mv1.ModelFormat

	hfModelRepos map[string]bool
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

// CreateHFModelRepo creates a new HuggingFace model repo.
func (c *FakeModelClient) CreateHFModelRepo(ctx context.Context, in *mv1.CreateHFModelRepoRequest, opts ...grpc.CallOption) (*mv1.HFModelRepo, error) {
	if _, ok := c.hfModelRepos[in.Name]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "hugging-face model repo %q already exists", in.Name)
	}
	c.hfModelRepos[in.Name] = true
	return &mv1.HFModelRepo{Name: in.Name}, nil

}

// GetHFModelRepo returns a HuggingFace model repo.
func (c *FakeModelClient) GetHFModelRepo(ctx context.Context, in *mv1.GetHFModelRepoRequest, opts ...grpc.CallOption) (*mv1.HFModelRepo, error) {
	if _, ok := c.hfModelRepos[in.Name]; !ok {
		return nil, status.Errorf(codes.NotFound, "hugging-face model repo %q not found", in.Name)
	}
	return &mv1.HFModelRepo{Name: in.Name}, nil
}

// New creates a new loader.
func New(
	objectStorePathPrefix string,
	baseModelPathPrefix string,
	modelDownloader ModelDownloader,
	isDownloadingFromHuggingFace bool,
	s3Client S3Client,
	modelClient ModelClient,
	log logr.Logger,
) *L {
	return &L{
		objectStorePathPrefix:        objectStorePathPrefix,
		baseModelPathPrefix:          baseModelPathPrefix,
		modelDownloader:              modelDownloader,
		isDownloadingFromHuggingFace: isDownloadingFromHuggingFace,
		s3Client:                     s3Client,
		modelClient:                  modelClient,
		log:                          log.WithName("loader"),
		tmpDir:                       "/tmp",
	}
}

// L is a loader.
type L struct {
	// objectStorePathPrefix is the prefix of the path to the base and non-base models in the object stoer.
	objectStorePathPrefix string
	baseModelPathPrefix   string

	modelDownloader              ModelDownloader
	isDownloadingFromHuggingFace bool

	s3Client S3Client

	modelClient ModelClient

	log logr.Logger

	tmpDir string
}

// Run runs the loader.
func (l *L) Run(
	ctx context.Context,
	baseModels []string,
	models []config.ModelConfig,
	interval time.Duration,
) error {
	if err := l.LoadModels(ctx, baseModels, models); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// TODO(kenji): Change this to dynamic model loading.
			if err := l.LoadModels(ctx, baseModels, models); err != nil {
				return err
			}
		}
	}
}

// LoadModels loads base and non-base models.
func (l *L) LoadModels(
	ctx context.Context,
	baseModels []string,
	models []config.ModelConfig,
) error {
	for _, baseModel := range baseModels {
		if err := l.loadBaseModel(ctx, baseModel); err != nil {
			return err
		}
	}
	for _, m := range models {
		if err := l.loadModel(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (l *L) loadBaseModel(ctx context.Context, modelID string) error {
	convertedModelID := toLLMarinerModelID(modelID)

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

	modelIDToDownload, filename, err := splitHFRepoAndFile(modelID)
	if err != nil {
		return err
	}

	// Check if the HF repo has already been downloaded. We need to check if when
	// a repo contains multiple GGUF files as there is no one-to-one mapping between the repo and
	// base models.
	if l.isDownloadingFromHuggingFace && filename == "" {
		// Check if the corresponding HuggingFace repo has already been downloaded.
		_, err := l.modelClient.GetHFModelRepo(ctx, &v1.GetHFModelRepoRequest{Name: modelID})
		if err == nil {
			l.log.Info("Already HuggingFace model repo exists", "modelID", modelID)
			return nil
		}
		if status.Code(err) != codes.NotFound {
			return err
		}
	}

	modelInfos, err := l.downloadAndUploadModel(ctx, modelIDToDownload, filename, true)
	if err != nil {
		return err
	}

	for _, mi := range modelInfos {
		modelID := toLLMarinerModelID(mi.id)
		_, err := l.modelClient.GetBaseModelPath(ctx, &mv1.GetBaseModelPathRequest{Id: modelID})
		if err == nil {
			l.log.Info("Already model exists", "modelID", modelID)
			continue
		}
		if status.Code(err) != codes.NotFound {
			return err
		}

		if _, err := l.modelClient.CreateBaseModel(ctx, &mv1.CreateBaseModelRequest{
			Id:            modelID,
			Path:          mi.path,
			Formats:       mi.formats,
			GgufModelPath: mi.ggufModelPath,
		}); err != nil {
			return err
		}

		l.log.Info("Successfully loaded base model", "model", modelID)
	}

	if l.isDownloadingFromHuggingFace && filename == "" {
		if _, err := l.modelClient.CreateHFModelRepo(ctx, &v1.CreateHFModelRepoRequest{Name: modelID}); err != nil {
			return err
		}
	}

	return nil
}

func (l *L) loadModel(ctx context.Context, model config.ModelConfig) error {
	if err := l.loadBaseModel(ctx, model.BaseModel); err != nil {
		return err
	}

	convertedModelID := toLLMarinerModelID(model.Model)

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

	modelInfos, err := l.downloadAndUploadModel(ctx, model.Model, "", false)
	if err != nil {
		return err
	}
	if n := len(modelInfos); n != 1 {
		return fmt.Errorf("unsupported number of model infos: %d", n)
	}
	mi := modelInfos[0]

	if _, err := l.modelClient.RegisterModel(ctx, &mv1.RegisterModelRequest{
		Id:           convertedModelID,
		BaseModel:    toLLMarinerModelID(model.BaseModel),
		Adapter:      config.ToAdapterType(model.AdapterType),
		Quantization: config.ToQuantizationType(model.QuantizationType),
		Path:         mi.path,
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
	id string
	// path is the base path of model files.
	path          string
	ggufModelPath string
	formats       []v1.ModelFormat
}

func (l *L) downloadAndUploadModel(ctx context.Context, modelID, filename string, isBaseModel bool) ([]*modelInfo, error) {
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
		return nil, err
	}
	log.Info("Created a temp dir", "path", tmpDir)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	log.Info("Downloading model")
	if err := l.modelDownloader.download(ctx, modelID, filename, tmpDir); err != nil {
		return nil, err
	}

	keyModelID := strings.ReplaceAll(modelID, ":", "-")
	toKey := func(path string) string {
		// Remove the tmpdir path from the path.
		relativePath := strings.TrimPrefix(path, tmpDir)
		return filepath.Join(l.toPathPrefix(isBaseModel), keyModelID, relativePath)
	}

	var (
		paths                    []string
		ggufModelPaths           []string
		hasHuggingFaceConfigJSON bool
		hasTensorRTLLMConfig     bool
		hasOllamaLayerDir        bool
	)
	if err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "blobs" {
				hasOllamaLayerDir = true
			}
			return nil
		}

		paths = append(paths, path)

		if strings.HasSuffix(path, ".gguf") {
			ggufModelPaths = append(ggufModelPaths, toKey(path))
		}

		if filepath.Base(path) == "config.json" || filepath.Base(path) == "adapter_config.json" {
			hasHuggingFaceConfigJSON = true
		}

		if strings.HasSuffix(path, "tensorrt_llm/config.pbtxt") {
			hasTensorRTLLMConfig = true
		}

		return nil
	}); err != nil {
		return nil, err
	}
	log.Info("Downloaded files", "count", len(paths))
	if len(paths) == 0 {
		return nil, fmt.Errorf("no files downloaded")
	}

	log.Info("Uploading model to the object store")
	for _, path := range paths {
		log.Info("Uploading", "path", path)
		r, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		if err := l.s3Client.Upload(ctx, r, toKey(path)); err != nil {
			return nil, err
		}
		if err := r.Close(); err != nil {
			return nil, err
		}
	}

	if hasOllamaLayerDir {
		return []*modelInfo{
			{
				id:      modelID,
				path:    filepath.Join(l.toPathPrefix(isBaseModel), keyModelID),
				formats: []v1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_OLLAMA},
			},
		}, nil
	}

	var formats []v1.ModelFormat
	if hasHuggingFaceConfigJSON {
		formats = append(formats, mv1.ModelFormat_MODEL_FORMAT_HUGGING_FACE)
	}
	if hasTensorRTLLMConfig {
		formats = append(formats, mv1.ModelFormat_MODEL_FORMAT_NVIDIA_TRITON)
	}

	if len(ggufModelPaths) <= 1 {
		var ggufModelPath string
		if len(ggufModelPaths) == 1 {
			formats = append(formats, mv1.ModelFormat_MODEL_FORMAT_GGUF)
			ggufModelPath = ggufModelPaths[0]
		}

		if len(formats) == 0 {
			return nil, fmt.Errorf("no model format found")
		}

		id := modelID
		if filename != "" {
			id = modelID + "/" + filename
		}
		return []*modelInfo{
			{
				id:            id,
				path:          filepath.Join(l.toPathPrefix(isBaseModel), modelID),
				ggufModelPath: ggufModelPath,
				formats:       formats,
			},
		}, nil
	}

	log.Info("Found multiple GGPU files. Creating a model info per GGUF file")
	if len(formats) != 0 {
		return nil, fmt.Errorf("multiple gguf files (%v) found with other model format (%v)", ggufModelPaths, formats)
	}

	var minfos []*modelInfo
	for _, gpath := range ggufModelPaths {
		id := buildModelIDForGGUF(modelID, gpath)

		minfos = append(minfos, &modelInfo{
			id:            id,
			path:          filepath.Join(l.toPathPrefix(isBaseModel), id),
			ggufModelPath: gpath,
			formats:       []v1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_GGUF},
		})
	}

	return minfos, nil
}

func (l *L) toPathPrefix(isBaseModel bool) string {
	if isBaseModel {
		return filepath.Join(l.objectStorePathPrefix, l.baseModelPathPrefix)
	}
	// TODO(guangrui): make tenant-id configurable. The path should match with the path generated in RegisterModel.
	return filepath.Join(l.objectStorePathPrefix, "default-tenant-id")
}

func buildModelIDForGGUF(modelID, ggufModelFilePath string) string {
	p := strings.TrimSuffix(ggufModelFilePath, ".gguf")
	return filepath.Join(modelID, filepath.Base(p))
}

func toLLMarinerModelID(id string) string {
	// HuggingFace uses '/' as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	return strings.ReplaceAll(id, "/", "-")
}

// splitHFRepoAndFile returns the HuggingFace repo name and the filename to be downloaded.
// If the modelID is of the form "repo/filename", the function returns "repo" and "filename".
func splitHFRepoAndFile(modelID string) (string, string, error) {
	l := strings.Split(modelID, "/")
	if len(l) > 3 {
		return "", "", fmt.Errorf("unexpected model ID format: %s", modelID)
	}
	if len(l) == 3 {
		return l[0] + "/" + l[1], l[2], nil
	}
	return modelID, "", nil
}
