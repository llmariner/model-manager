package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	"github.com/llmariner/common/pkg/id"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// CreateModel creates a base model.
func (s *S) CreateModel(
	ctx context.Context,
	req *v1.CreateModelRequest,
) (*v1.Model, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	// TODO(kenji): Revisit the permission check. The base model is scoped by a tenant, not project,

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.SourceRepository == v1.SourceRepository_SOURCE_REPOSITORY_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "source_repository is required")
	}

	if req.SourceRepository != v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE &&
		req.SourceRepository != v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE &&
		req.SourceRepository != v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA {
		return nil, status.Errorf(codes.InvalidArgument, "source_repository must be one of %v", []v1.SourceRepository{
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA,
		})
	}

	m, err := s.store.CreateBaseModelWithLoadingRequested(req.Id, req.SourceRepository, userInfo.TenantID)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "create base model: %s", err)
	}

	return baseToModelProto(m), nil
}

// ListModels lists models.
func (s *S) ListModels(
	ctx context.Context,
	req *v1.ListModelsRequest,
) (*v1.ListModelsResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	var modelProtos []*v1.Model
	// First include base models.
	bms, err := s.store.ListBaseModels(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}
	for _, m := range bms {
		if !isBaseModelLoaded(m) && !req.IncludeLoadingModels {
			continue
		}

		modelProtos = append(modelProtos, baseToModelProto(m))
	}

	// Then add generated models owned by the project
	ms, err := s.store.ListModelsByProjectID(userInfo.ProjectID, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}

	for _, m := range ms {
		modelProtos = append(modelProtos, toModelProto(m))
	}

	return &v1.ListModelsResponse{
		Object: "list",
		Data:   modelProtos,
	}, nil
}

// GetModel gets a model.
func (s *S) GetModel(
	ctx context.Context,
	req *v1.GetModelRequest,
) (*v1.Model, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// First check if it's a base model.
	bm, err := s.store.GetBaseModel(req.Id, userInfo.TenantID)
	if err == nil {
		if !isBaseModelLoaded(bm) && !req.IncludeLoadingModel {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}

		return baseToModelProto(bm), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	// Then check if it's a generated model.
	m, err := s.store.GetPublishedModelByModelIDAndProjectID(req.Id, userInfo.ProjectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get published model by model id and tenant id: %s", err)
	}
	return toModelProto(m), nil
}

// DeleteModel deletes a model.
func (s *S) DeleteModel(
	ctx context.Context,
	req *v1.DeleteModelRequest,
) (*v1.DeleteModelResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if _, err := s.store.GetBaseModel(req.Id, userInfo.TenantID); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}

		// The specified model is not a base-model or the base-model has already been deleted.
		if err := s.store.DeleteModel(req.Id, userInfo.ProjectID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
			}
			return nil, status.Errorf(codes.Internal, "delete model: %s", err)
		}

		return &v1.DeleteModelResponse{
			Id:      req.Id,
			Object:  "model",
			Deleted: true,
		}, nil
	}

	// The specified model is a base-model. Delete it.
	//
	// TODO(kenji): Revisit the permission check. The base model is scoped by a tenant, not project,
	// so we should have additional check here.
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteBaseModelInTransaction(tx, req.Id, userInfo.TenantID); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.Internal, "delete model: %s", err)
			}
			return status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}

		// Delete the HFModelRepo if the model is from Hugging Face. Otherwise the same
		// model cannot be reloaded again.
		//
		// TODO(kenji): Handle a case where a single Hugging Face repo has multiple models. In that case,
		// the Hugging Face repo name and the model ID does not match.
		//
		// Also, deleting a HFModelRepo can trigger downloading the remaining undeleted models again, which is not ideal.

		// Replace the first "-" with "-" to have the original repo name.
		hfRepoName := strings.Replace(req.Id, "-", "/", 1)
		if err := store.DeleteHFModelRepoInTransaction(tx, hfRepoName, userInfo.TenantID); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.Internal, "delete hf model repo (id: %q, repo name: %q): %s", req.Id, hfRepoName, err)
			}
			// Ignore. The HFModelRepo does not exist for old models or non-HF models.
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &v1.DeleteModelResponse{
		Id:      req.Id,
		Object:  "model",
		Deleted: true,
	}, nil
}

// ListBaseModels lists base models.
func (s *S) ListBaseModels(
	ctx context.Context,
	req *v1.ListBaseModelsRequest,
) (*v1.ListBaseModelsResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	ms, err := s.store.ListBaseModels(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list base models: %s", err)
	}
	var modelProtos []*v1.BaseModel
	for _, m := range ms {
		modelProtos = append(modelProtos, toBaseModelProto(m))
	}
	return &v1.ListBaseModelsResponse{
		Object: "list",
		Data:   modelProtos,
	}, nil
}

// GetModel gets a model.
func (s *WS) GetModel(
	ctx context.Context,
	req *v1.GetModelRequest,
) (*v1.Model, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	// First check if it's a base model.
	bm, err := s.store.GetBaseModel(req.Id, clusterInfo.TenantID)
	if err == nil {
		return baseToModelProto(bm), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	// Then check if it's a generated model.
	m, err := s.store.GetPublishedModelByModelIDAndTenantID(req.Id, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get published model by model id and tenant id: %s", err)
	}
	return toModelProto(m), nil
}

// RegisterModel registers a model.
func (s *WS) RegisterModel(
	ctx context.Context,
	req *v1.RegisterModelRequest,
) (*v1.RegisterModelResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if req.Id == "" && req.Suffix == "" {
		return nil, status.Error(codes.InvalidArgument, "id or suffix is required")
	}
	if req.BaseModel == "" {
		return nil, status.Error(codes.InvalidArgument, "base_model is required")
	}
	if req.OrganizationId == "" {
		return nil, status.Error(codes.InvalidArgument, "organization_id is required")
	}
	if req.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id is required")
	}

	id := req.Id
	if id != "" {
		_, err := s.store.GetModelByModelID(id)
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", id)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get model by model ID: %s", err)
		}
	} else {
		id, err = s.genenerateModelID(req.BaseModel, req.Suffix)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "generate model ID: %s", err)
		}
	}

	path := req.Path
	if path == "" {
		sc, err := s.store.GetStorageConfig(clusterInfo.TenantID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get storage config: %s", err)
		}
		path = fmt.Sprintf("%s/%s/%s", sc.PathPrefix, clusterInfo.TenantID, id)
	}
	_, err = s.store.CreateModel(store.ModelSpec{
		ModelID:        id,
		TenantID:       clusterInfo.TenantID,
		OrganizationID: req.OrganizationId,
		ProjectID:      req.ProjectId,
		IsPublished:    false,
		Path:           path,
		BaseModelID:    req.BaseModel,
		Adapter:        req.Adapter,
		Quantization:   req.Quantization,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}

	return &v1.RegisterModelResponse{
		Id:   id,
		Path: path,
	}, nil
}

// PublishModel publishes a model.
func (s *WS) PublishModel(
	ctx context.Context,
	req *v1.PublishModelRequest,
) (*v1.PublishModelResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.store.UpdateModel(req.Id, clusterInfo.TenantID, true); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "update model: %s", err)
	}
	return &v1.PublishModelResponse{}, nil
}

// GetModelPath gets a model path.
func (s *WS) GetModelPath(
	ctx context.Context,
	req *v1.GetModelPathRequest,
) (*v1.GetModelPathResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	m, err := s.store.GetPublishedModelByModelIDAndTenantID(req.Id, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get published model by model id and tenant id: %s", err)
	}
	return &v1.GetModelPathResponse{
		Path: m.Path,
	}, nil
}

// GetModelAttributes gets the model attributes.
func (s *WS) GetModelAttributes(
	ctx context.Context,
	req *v1.GetModelAttributesRequest,
) (*v1.ModelAttributes, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	m, err := s.store.GetPublishedModelByModelIDAndTenantID(req.Id, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get published model by model id and tenant id: %s", err)
	}
	return &v1.ModelAttributes{
		Path:         m.Path,
		BaseModel:    m.BaseModelID,
		Adapter:      m.Adapter,
		Quantization: m.Quantization,
	}, nil
}

// CreateBaseModel creates a base model.
func (s *WS) CreateBaseModel(
	ctx context.Context,
	req *v1.CreateBaseModelRequest,
) (*v1.BaseModel, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Path == "" {
		return nil, status.Error(codes.InvalidArgument, "path is required")
	}

	var formats []v1.ModelFormat
	if len(req.Formats) == 0 {
		// Fall back to GGUF for backward compatibility.
		// TODO(kenji): Make this to return an errror once all clients are updated.
		formats = append(formats, v1.ModelFormat_MODEL_FORMAT_GGUF)
	} else {
		formats = append(formats, req.Formats...)
	}

	for _, format := range formats {
		if format == v1.ModelFormat_MODEL_FORMAT_GGUF && req.GgufModelPath == "" {
			return nil, status.Error(codes.InvalidArgument, "gguf_model_path is required for the GGUF format")
		}
	}

	// Note: We skip the validation of source repository for backward compatibility.

	existing, err := s.store.GetBaseModel(req.Id, clusterInfo.TenantID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}

		// Create a new base model.

		m, err := s.store.CreateBaseModel(req.Id, req.Path, formats, req.GgufModelPath, req.SourceRepository, clusterInfo.TenantID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "create base model: %s", err)
		}
		return toBaseModelProto(m), nil
	}

	if isBaseModelLoaded(existing) {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", req.Id)
	}

	// Update the existing model.
	if err := s.store.UpdateBaseModelToSucceededStatus(
		req.Id,
		clusterInfo.TenantID,
		req.Path,
		formats,
		req.GgufModelPath,
	); err != nil {
		return nil, status.Errorf(codes.Internal, "update base model: %s", err)
	}
	return toBaseModelProto(existing), nil
}

// GetBaseModelPath gets a model path.
func (s *WS) GetBaseModelPath(
	ctx context.Context,
	req *v1.GetBaseModelPathRequest,
) (*v1.GetBaseModelPathResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	m, err := s.store.GetBaseModel(req.Id, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	if !isBaseModelLoaded(m) {
		return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
	}

	formats, err := store.UnmarshalModelFormats(m.Formats)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unmarshal model formats: %s", err)
	}

	// Add GGUF as format if GGUF model path is set (for backward compatibility).
	if m.GGUFModelPath != "" && len(formats) == 0 {
		formats = append(formats, v1.ModelFormat_MODEL_FORMAT_GGUF)
	}

	return &v1.GetBaseModelPathResponse{
		Path:          m.Path,
		Formats:       formats,
		GgufModelPath: m.GGUFModelPath,
	}, nil
}

// AcquireUnloadedBaseModel checks if there is any unloaded base model. If exists,
// update the loading status to LOADED and return it.
func (s *WS) AcquireUnloadedBaseModel(
	ctx context.Context,
	req *v1.AcquireUnloadedBaseModelRequest,
) (*v1.AcquireUnloadedBaseModelResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ms, err := s.store.ListUnloadedBaseModels(clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list unloaded base models: %s", err)
	}

	if len(ms) == 0 {
		return &v1.AcquireUnloadedBaseModelResponse{}, nil
	}

	m := ms[0]
	if err := s.store.UpdateBaseModelToLoadingStatus(m.ModelID, clusterInfo.TenantID); err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update base model loading status: %s", err)
	}

	return &v1.AcquireUnloadedBaseModelResponse{
		BaseModelId:      m.ModelID,
		SourceRepository: m.SourceRepository,
	}, nil
}

// UpdateBaseModelLoadingStatus updates the loading status. When the loading succeeded, it also
// updates the base model metadata.
func (s *WS) UpdateBaseModelLoadingStatus(
	ctx context.Context,
	req *v1.UpdateBaseModelLoadingStatusRequest,
) (*v1.UpdateBaseModelLoadingStatusResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.LoadingResult == nil {
		return nil, status.Error(codes.InvalidArgument, "loading_result is required")
	}

	bm, err := s.store.GetBaseModel(req.Id, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	switch req.LoadingResult.(type) {
	case *v1.UpdateBaseModelLoadingStatusRequest_Success_:
		// model-manager-loader calls this RPC after making the CreateBaseModel RPC request.
		//
		// If the model's loading status is still Loading, this indicates either
		// model ID mismatch due to our internal conversion (e.g., "/" to "-") or
		// a Hugging Face repo contains multiple models.
		//
		// In this case, we should delete delete the requested model ID.
		if bm.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING {
			s.log.Info("Delete the model %q as base models have been successfully created", "modelID", req.Id)
			if err := s.store.DeleteBaseModel(req.Id, clusterInfo.TenantID); err != nil {
				return nil, status.Errorf(codes.Internal, "delete base model: %s", err)
			}
		}
	case *v1.UpdateBaseModelLoadingStatusRequest_Failure_:
		failure := req.GetFailure()
		if failure.Reason == "" {
			return nil, status.Error(codes.InvalidArgument, "reason is required")
		}
		err = s.store.UpdateBaseModelToFailedStatus(
			req.Id,
			clusterInfo.TenantID,
			failure.Reason,
		)
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid loading_result")
	}

	if err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update base model loading status: %s", err)
	}

	return &v1.UpdateBaseModelLoadingStatusResponse{}, nil
}

func (s *WS) genenerateModelID(baseModel, suffix string) (string, error) {
	const randomLength = 10
	// OpenAI uses ':" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	// Replace "/" with "-'. HuggingFace model contains "/", but that doesn't work for Ollama.
	base := fmt.Sprintf("ft:%s:%s-", strings.ReplaceAll(baseModel, "/", "-"), suffix)

	// Randomly create an ID and retry if it already exists.
	for {
		randomStr, err := id.GenerateID("", randomLength)
		if err != nil {
			return "", fmt.Errorf("generate ID: %s", err)
		}
		id := fmt.Sprintf("%s%s", base, randomStr)
		if _, err := s.store.GetModelByModelID(id); errors.Is(err, gorm.ErrRecordNotFound) {
			return id, nil
		}
	}
}

func toModelProto(m *store.Model) *v1.Model {
	return &v1.Model{
		Id:               m.ModelID,
		Object:           "model",
		Created:          m.CreatedAt.UTC().Unix(),
		OwnedBy:          "user",
		LoadingStatus:    v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_FINE_TUNING,
	}
}

func baseToModelProto(m *store.BaseModel) *v1.Model {
	return &v1.Model{
		Id:                   m.ModelID,
		Object:               "model",
		Created:              m.CreatedAt.UTC().Unix(),
		OwnedBy:              "system",
		LoadingStatus:        m.LoadingStatus,
		SourceRepository:     m.SourceRepository,
		LoadingFailureReason: m.LoadingFailureReason,
	}
}

func toBaseModelProto(m *store.BaseModel) *v1.BaseModel {
	return &v1.BaseModel{
		Id:      m.ModelID,
		Object:  "basemodel",
		Created: m.CreatedAt.UTC().Unix(),
	}
}

func isBaseModelLoaded(m *store.BaseModel) bool {
	// The UNSPECIFIED status is considered as loaded for backward compatibility.
	return m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED || m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_UNSPECIFIED
}
