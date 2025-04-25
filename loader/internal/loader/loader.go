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
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/loader/internal/config"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ModelDownloader is an interface for downloading a model.
type ModelDownloader interface {
	download(ctx context.Context, modelPath, filename, destDir string) error
}

// modelDownloaderFactory is the factory for ModelDownloader.
type modelDownloaderFactory interface {
	Create(context.Context, v1.SourceRepository) (ModelDownloader, error)
}

// S3Client is an interface for uploading a file to S3.
type S3Client interface {
	Upload(ctx context.Context, r io.Reader, bucket, key string) error
}

// NoopS3Client is a no-op S3 client.
type NoopS3Client struct {
}

// Upload uploads a file to S3.
func (c *NoopS3Client) Upload(ctx context.Context, r io.Reader, bucket, key string) error {
	return nil
}

// ModelClient is an interface for the model client.
type ModelClient interface {
	CreateBaseModel(ctx context.Context, in *v1.CreateBaseModelRequest, opts ...grpc.CallOption) (*v1.BaseModel, error)
	GetBaseModelPath(ctx context.Context, in *v1.GetBaseModelPathRequest, opts ...grpc.CallOption) (*v1.GetBaseModelPathResponse, error)
	GetModelPath(ctx context.Context, in *v1.GetModelPathRequest, opts ...grpc.CallOption) (*v1.GetModelPathResponse, error)
	RegisterModel(ctx context.Context, in *v1.RegisterModelRequest, opts ...grpc.CallOption) (*v1.RegisterModelResponse, error)
	PublishModel(ctx context.Context, in *v1.PublishModelRequest, opts ...grpc.CallOption) (*v1.PublishModelResponse, error)

	CreateHFModelRepo(ctx context.Context, in *v1.CreateHFModelRepoRequest, opts ...grpc.CallOption) (*v1.HFModelRepo, error)
	GetHFModelRepo(ctx context.Context, in *v1.GetHFModelRepoRequest, opts ...grpc.CallOption) (*v1.HFModelRepo, error)

	AcquireUnloadedBaseModel(ctx context.Context, in *v1.AcquireUnloadedBaseModelRequest, opts ...grpc.CallOption) (*v1.AcquireUnloadedBaseModelResponse, error)
	UpdateBaseModelLoadingStatus(ctx context.Context, in *v1.UpdateBaseModelLoadingStatusRequest, opts ...grpc.CallOption) (*v1.UpdateBaseModelLoadingStatusResponse, error)

	AcquireUnloadedModel(ctx context.Context, in *v1.AcquireUnloadedModelRequest, opts ...grpc.CallOption) (*v1.AcquireUnloadedModelResponse, error)
	UpdateModelLoadingStatus(ctx context.Context, in *v1.UpdateModelLoadingStatusRequest, opts ...grpc.CallOption) (*v1.UpdateModelLoadingStatusResponse, error)
}

// New creates a new loader.
func New(
	objectStoreBucket string,
	objectStorePathPrefix string,
	baseModelPathPrefix string,
	modelDownloaderFactory modelDownloaderFactory,
	s3Client S3Client,
	modelClient ModelClient,
	log logr.Logger,
) *L {
	return &L{
		objectStoreBucket:      objectStoreBucket,
		objectStorePathPrefix:  objectStorePathPrefix,
		baseModelPathPrefix:    baseModelPathPrefix,
		modelDownloaderFactory: modelDownloaderFactory,
		s3Client:               s3Client,
		modelClient:            modelClient,
		log:                    log.WithName("loader"),
		tmpDir:                 "/tmp",
	}
}

// L is a loader.
type L struct {
	objectStoreBucket string
	// objectStorePathPrefix is the prefix of the path to the base and non-base models in the object stoer.
	objectStorePathPrefix string
	baseModelPathPrefix   string

	modelDownloaderFactory modelDownloaderFactory

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
	sourceRepository v1.SourceRepository,
	interval time.Duration,
) error {
	if err := l.LoadModels(ctx, baseModels, models, sourceRepository); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := l.pullAndLoadBaseModels(ctx); err != nil {
				return err
			}
			if err := l.pullAndLoadModels(ctx); err != nil {
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
	sourceRepository v1.SourceRepository,
) error {
	for _, baseModel := range baseModels {
		if err := l.loadBaseModel(ctx, baseModel, sourceRepository); err != nil {
			return err
		}
	}
	for _, m := range models {
		if err := l.loadModelFromConfig(ctx, m, sourceRepository); err != nil {
			return err
		}
	}
	return nil
}

func (l *L) pullAndLoadBaseModels(ctx context.Context) error {
	actx := auth.AppendWorkerAuthorization(ctx)
	for {
		resp, err := l.modelClient.AcquireUnloadedBaseModel(actx, &v1.AcquireUnloadedBaseModelRequest{})
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				l.log.Error(err, "Failed to acquire an unloaded base model")
				continue
			}
			return err
		}

		if resp.BaseModelId == "" {
			l.log.Info("No unloaded base model")
			return nil
		}

		if err := l.loadBaseModel(ctx, resp.BaseModelId, resp.SourceRepository); err != nil {
			l.log.Error(err, "Failed to load base model", "modelID", resp.BaseModelId)
			if _, err := l.modelClient.UpdateBaseModelLoadingStatus(actx, &v1.UpdateBaseModelLoadingStatusRequest{
				Id: resp.BaseModelId,
				LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Failure_{
					Failure: &v1.UpdateBaseModelLoadingStatusRequest_Failure{
						Reason: err.Error(),
					},
				},
			}); err != nil {
				return err
			}
			// Do not return the error here. We need to continue loading other models.
			continue
		}

		l.log.Info("Successfully loaded base model", "modelID", resp.BaseModelId)
		if _, err := l.modelClient.UpdateBaseModelLoadingStatus(actx, &v1.UpdateBaseModelLoadingStatusRequest{
			Id:            resp.BaseModelId,
			LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Success_{},
		}); err != nil {
			return err
		}
	}
}

func (l *L) pullAndLoadModels(ctx context.Context) error {
	actx := auth.AppendWorkerAuthorization(ctx)
	for {
		resp, err := l.modelClient.AcquireUnloadedModel(actx, &v1.AcquireUnloadedModelRequest{})
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				l.log.Error(err, "Failed to acquire an unloaded base model")
				continue
			}
			return err
		}

		if resp.ModelId == "" {
			l.log.Info("No unloaded model")
			return nil
		}

		if _, err := l.downloadAndUploadModel(ctx, resp.ModelId, resp.ModelFileLocation, "", resp.SourceRepository, resp.DestPath); err != nil {
			l.log.Error(err, "Failed to load model", "modelID", resp.ModelId)
			if _, err := l.modelClient.UpdateModelLoadingStatus(actx, &v1.UpdateModelLoadingStatusRequest{
				Id: resp.ModelId,
				LoadingResult: &v1.UpdateModelLoadingStatusRequest_Failure_{
					Failure: &v1.UpdateModelLoadingStatusRequest_Failure{
						Reason: err.Error(),
					},
				},
			}); err != nil {
				return err
			}
			// Do not return the error here. We need to continue loading other models.
			continue
		}

		l.log.Info("Successfully loaded base model", "modelID", resp.ModelId)
		if _, err := l.modelClient.UpdateModelLoadingStatus(actx, &v1.UpdateModelLoadingStatusRequest{
			Id:            resp.ModelId,
			LoadingResult: &v1.UpdateModelLoadingStatusRequest_Success_{},
		}); err != nil {
			return err
		}
	}
}

func (l *L) loadBaseModel(ctx context.Context, modelID string, sourceRepository v1.SourceRepository) error {
	convertedModelID := toLLMarinerModelID(modelID)

	// First check if the model exists in the database.
	ctx = auth.AppendWorkerAuthorization(ctx)
	_, err := l.modelClient.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{Id: convertedModelID})
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
	isDownloadingFromHuggingFace := sourceRepository == v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE
	if isDownloadingFromHuggingFace && filename == "" {
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

	pathPrefix := filepath.Join(l.objectStorePathPrefix, l.baseModelPathPrefix, toKeyModelID(modelIDToDownload))
	modelInfos, err := l.downloadAndUploadModel(ctx, modelIDToDownload, modelIDToDownload, filename, sourceRepository, pathPrefix)
	if err != nil {
		return err
	}

	for _, mi := range modelInfos {
		modelID := toLLMarinerModelID(mi.id)
		_, err := l.modelClient.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{Id: modelID})
		if err == nil {
			l.log.Info("Already model exists", "modelID", modelID)
			continue
		}
		if status.Code(err) != codes.NotFound {
			return err
		}

		if _, err := l.modelClient.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
			Id:               modelID,
			Path:             mi.path,
			Formats:          mi.formats,
			GgufModelPath:    mi.ggufModelPath,
			SourceRepository: mi.sourceRepository,
		}); err != nil {
			return err
		}

		l.log.Info("Successfully loaded base model", "model", modelID)
	}

	if isDownloadingFromHuggingFace && filename == "" {
		if _, err := l.modelClient.CreateHFModelRepo(ctx, &v1.CreateHFModelRepoRequest{Name: modelID}); err != nil {
			return err
		}
	}

	return nil
}

func (l *L) loadModelFromConfig(ctx context.Context, model config.ModelConfig, sourceRepository v1.SourceRepository) error {
	// TODO(guangrui): Make these configurable.
	const (
		projectID      = "default"
		organizationID = "default"
		tenantID       = "default-tenant-id"
	)
	if err := l.loadBaseModel(ctx, model.BaseModel, sourceRepository); err != nil {
		return err
	}

	convertedModelID := toLLMarinerModelID(model.Model)

	// First check if the model exists in the database.
	ctx = auth.AppendWorkerAuthorization(ctx)
	_, err := l.modelClient.GetModelPath(ctx, &v1.GetModelPathRequest{Id: convertedModelID})
	if err == nil {
		l.log.Info("Already model exists", "modelID", convertedModelID)
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return err
	}

	// TODO(guangrui): make tenant-id configurable. The path should match with the path generated in RegisterModel.
	pathPrefix := filepath.Join(l.objectStorePathPrefix, tenantID, toKeyModelID(model.Model))
	modelInfos, err := l.downloadAndUploadModel(ctx, model.Model, model.Model, "", sourceRepository, pathPrefix)
	if err != nil {
		return err
	}
	if n := len(modelInfos); n != 1 {
		return fmt.Errorf("unsupported number of model infos: %d", n)
	}
	mi := modelInfos[0]

	if _, err := l.modelClient.RegisterModel(ctx, &v1.RegisterModelRequest{
		Id:           convertedModelID,
		BaseModel:    toLLMarinerModelID(model.BaseModel),
		Adapter:      config.ToAdapterType(model.AdapterType),
		Quantization: config.ToQuantizationType(model.QuantizationType),
		Path:         mi.path,

		ProjectId:      projectID,
		OrganizationId: organizationID,
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
	path             string
	ggufModelPath    string
	formats          []v1.ModelFormat
	sourceRepository v1.SourceRepository
}

func (l *L) downloadAndUploadModel(
	ctx context.Context,
	modelID,
	modelPath,
	filename string,
	sourceRepository v1.SourceRepository,
	pathPrefix string,
) ([]*modelInfo, error) {
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
	downloader, err := l.modelDownloaderFactory.Create(ctx, sourceRepository)
	if err != nil {
		return nil, err
	}
	if err := downloader.download(ctx, modelPath, filename, tmpDir); err != nil {
		return nil, err
	}

	toKey := func(path string) string {
		// Remove the tmpdir path from the path.
		relativePath := strings.TrimPrefix(path, tmpDir)
		return filepath.Join(pathPrefix, relativePath)
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

		if err := l.s3Client.Upload(ctx, r, l.objectStoreBucket, toKey(path)); err != nil {
			return nil, err
		}
		if err := r.Close(); err != nil {
			return nil, err
		}
	}

	if hasOllamaLayerDir {
		return []*modelInfo{
			{
				id:               modelID,
				path:             pathPrefix,
				formats:          []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_OLLAMA},
				sourceRepository: sourceRepository,
			},
		}, nil
	}

	var formats []v1.ModelFormat
	if hasHuggingFaceConfigJSON {
		formats = append(formats, v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE)
	}
	if hasTensorRTLLMConfig {
		formats = append(formats, v1.ModelFormat_MODEL_FORMAT_NVIDIA_TRITON)
	}

	if len(ggufModelPaths) <= 1 {
		var ggufModelPath string
		if len(ggufModelPaths) == 1 {
			formats = append(formats, v1.ModelFormat_MODEL_FORMAT_GGUF)
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
				id:               id,
				path:             pathPrefix,
				ggufModelPath:    ggufModelPath,
				formats:          formats,
				sourceRepository: sourceRepository,
			},
		}, nil
	}

	log.Info("Found multiple GGPU files. Creating a model info per GGUF file")
	if len(formats) != 0 {
		return nil, fmt.Errorf("multiple gguf files (%v) found with other model format (%v)", ggufModelPaths, formats)
	}

	var minfos []*modelInfo
	for _, gpath := range ggufModelPaths {
		filename := extractFileNameFromGGUFPath(gpath)
		id := filepath.Join(modelID, filename)

		minfos = append(minfos, &modelInfo{
			id:               id,
			path:             filepath.Join(pathPrefix, filename),
			ggufModelPath:    gpath,
			formats:          []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
			sourceRepository: sourceRepository,
		})
	}

	return minfos, nil
}

func extractFileNameFromGGUFPath(path string) string {
	return filepath.Base(strings.TrimSuffix(path, ".gguf"))
}

func toKeyModelID(modelID string) string {
	// Ollama uses ':' as a separator, but it cannot be used for bucket name. Use '-' instead.
	return strings.ReplaceAll(modelID, ":", "-")
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
