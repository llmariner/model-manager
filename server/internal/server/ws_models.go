package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/llmariner/common/pkg/id"
	v1 "github.com/llmariner/model-manager/api/v1"
	mid "github.com/llmariner/model-manager/common/pkg/id"
	"github.com/llmariner/model-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ListModels lists models.
//
// TODO(kenji): Exclude models that shouldn't be loaded in the requesting cluster based on their model config.
func (s *WS) ListModels(ctx context.Context, req *v1.ListModelsRequest) (*v1.ListModelsResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bms, err := s.store.ListBaseModelsByTenantID(clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list base models: %s", err)
	}

	// When the same base model exists for multiple projects, we return all of them.
	// A caller (e.g., inference-manager-engine) decides which one to be loaded.
	var modelProtos []*v1.Model
	for _, m := range bms {
		if !isBaseModelLoaded(m) {
			continue
		}

		mp, err := convertBaseModelToProto(s.store, s.pcache, m)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "base model to proto: %s", err)
		}

		modelProtos = append(modelProtos, mp)
	}

	ms, err := s.store.ListModelsByTenantID(clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}

	for _, m := range ms {
		if !isModelLoaded(m) {
			continue
		}

		if !m.IsPublished {
			continue
		}

		mp, err := convertFineTunedModelToProto(s.store, s.pcache, m)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "model to proto: %s", err)
		}
		modelProtos = append(modelProtos, mp)
	}

	return &v1.ListModelsResponse{
		Object:     "list",
		Data:       modelProtos,
		HasMore:    false,
		TotalItems: int32(len(modelProtos)),
	}, nil
}

// RegisterModel registers a fine-tuned model.
// The model is created in the database, but not published yet.
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
		_, err := s.store.GetModelByModelIDAndTenantID(id, clusterInfo.TenantID)
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", id)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get model by model ID: %s", err)
		}
	} else {
		id, err = s.generateModelID(req.BaseModel, req.Suffix, clusterInfo.TenantID)
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
		path = fmt.Sprintf("%s/%s/%s/%s", sc.PathPrefix, clusterInfo.TenantID, req.ProjectId, id)
	}

	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if _, err := store.CreateModelInTransaction(tx, store.ModelSpec{
			ModelID:        id,
			TenantID:       clusterInfo.TenantID,
			OrganizationID: req.OrganizationId,
			ProjectID:      req.ProjectId,
			IsPublished:    false,
			Path:           path,
			BaseModelID:    req.BaseModel,
			Adapter:        req.Adapter,
			Quantization:   req.Quantization,
			LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		}); err != nil {
			return status.Errorf(codes.Internal, "create model: %s", err)
		}

		if err := store.CreateModelActivationStatusInTransaction(tx, &store.ModelActivationStatus{
			ModelID:  id,
			TenantID: clusterInfo.TenantID,
			Status:   v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
		}); err != nil {
			return status.Errorf(codes.Internal, "create model activation status: %s", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &v1.RegisterModelResponse{
		Id:   id,
		Path: path,
	}, nil
}

// PublishModel publishes a fine-tuned model.
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

	if err := s.store.UpdateModelPublishingStatus(req.Id, clusterInfo.TenantID, true, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED); err != nil {
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
	k := store.ModelKey{
		ModelID:   req.Id,
		ProjectID: req.ProjectId,
		TenantID:  clusterInfo.TenantID,
	}
	existing, err := s.store.GetBaseModel(k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}

		// Find the original model that is in the loading status.
		bms, err := s.store.ListLoadingBaseModels(clusterInfo.TenantID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "list loading base models: %s", err)
		}
		var originalBaseModel *store.BaseModel
		for _, m := range bms {
			if m.ProjectID != req.ProjectId {
				continue
			}
			convertedModelID := mid.ToLLMarinerModelID(m.ModelID)
			if convertedModelID == req.Id {
				originalBaseModel = m
				break
			}
			if strings.HasPrefix(req.Id, convertedModelID+"-") {
				// Handle a case where a new model is created for each file in a Hugging Face model repository.
				originalBaseModel = m
				break
			}
		}

		// Create a new base model.
		var m *store.BaseModel
		if err := s.store.Transaction(func(tx *gorm.DB) error {
			var err error
			m, err = store.CreateBaseModelInTransaction(tx, k, req.Path, formats, req.GgufModelPath, req.SourceRepository)
			if err != nil {
				return status.Errorf(codes.Internal, "create base model: %s", err)
			}

			if err := store.CreateModelActivationStatusInTransaction(tx, &store.ModelActivationStatus{
				ModelID:   k.ModelID,
				ProjectID: k.ProjectID,
				TenantID:  k.TenantID,
				Status:    v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
			}); err != nil {
				return status.Errorf(codes.Internal, "create model activation status: %s", err)
			}

			// Copy the model config from the original base model if it exists.
			if originalBaseModel != nil {
				existing, err := s.store.GetModelConfig(store.ModelKey{
					ModelID:   originalBaseModel.ModelID,
					ProjectID: originalBaseModel.ProjectID,
					TenantID:  originalBaseModel.TenantID,
				})
				if err != nil {
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return status.Errorf(codes.Internal, "get model config: %s", err)
					}
				} else {
					s.log.Info("Copying model config from existing base model to new base model",
						"originalBaseModelID", originalBaseModel.ModelID,
						"newBaseModelID", k.ModelID,
					)
					if err := store.CreateModelConfigInTransaction(tx, &store.ModelConfig{
						ModelID:       k.ModelID,
						ProjectID:     k.ProjectID,
						TenantID:      k.TenantID,
						EncodedConfig: existing.EncodedConfig,
					}); err != nil {
						return status.Errorf(codes.Internal, "create model config: %s", err)
					}
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}

		return toBaseModelProto(m), nil
	}

	if isBaseModelLoaded(existing) {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", req.Id)
	}

	// Update the existing model.
	if err := s.store.UpdateBaseModelToSucceededStatus(
		k,
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

	var m *store.BaseModel
	if req.ProjectId != "" {
		var err error
		m, err = s.store.GetBaseModel(store.ModelKey{
			ModelID:   req.Id,
			ProjectID: req.ProjectId,
			TenantID:  clusterInfo.TenantID,
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
			}
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}
		if !isBaseModelLoaded(m) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
	} else {
		var found bool
		m, found, err = s.getLoadedBaseModel(req.Id, clusterInfo.TenantID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get loaded base model: %s", err)
		}
		if !found {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
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
//
// TODO(kenji): Exclude models that shouldn't be loaded in the requesting cluster based on their model config.
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
	k := store.ModelKey{
		ModelID:   m.ModelID,
		ProjectID: m.ProjectID,
		TenantID:  m.TenantID,
	}
	if err := s.store.UpdateBaseModelToLoadingStatus(k); err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update base model loading status: %s", err)
	}

	return &v1.AcquireUnloadedBaseModelResponse{
		BaseModelId:      m.ModelID,
		SourceRepository: m.SourceRepository,
		ProjectId:        m.ProjectID,
	}, nil
}

// AcquireUnloadedModel checks if there is any unloaded model. If exists,
// update the loading status to LOADED and return it.
//
// TODO(kenji): Exclude models that shouldn't be loaded in the requesting cluster based on their model config.
func (s *WS) AcquireUnloadedModel(
	ctx context.Context,
	req *v1.AcquireUnloadedModelRequest,
) (*v1.AcquireUnloadedModelResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ms, err := s.store.ListUnloadedModels(clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list unloaded models: %s", err)
	}

	if len(ms) == 0 {
		return &v1.AcquireUnloadedModelResponse{}, nil
	}

	m := ms[0]
	if err := s.store.UpdateModelToLoadingStatus(m.ModelID, clusterInfo.TenantID); err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update model loading status: %s", err)
	}

	return &v1.AcquireUnloadedModelResponse{
		ModelId:           m.ModelID,
		IsBaseModel:       false,
		SourceRepository:  m.SourceRepository,
		ModelFileLocation: m.ModelFileLocation,
		DestPath:          m.Path,
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

	if req.LoadingResult == nil && req.StatusMessage == "" {
		return nil, status.Error(codes.InvalidArgument, "loading_result or status_message is required")
	}

	k := store.ModelKey{
		ModelID:   req.Id,
		ProjectID: req.ProjectId,
		TenantID:  clusterInfo.TenantID,
	}

	bm, err := s.store.GetBaseModel(k)
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
		// In this case, we should delete the requested model ID.
		if bm.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING {
			s.log.Info("Delete the model as a new base model has been successfully created", "modelID", req.Id)
			if err := s.store.Transaction(func(tx *gorm.DB) error {
				if err := store.DeleteBaseModelInTransaction(tx, k); err != nil {
					return status.Errorf(codes.Internal, "delete model: %s", err)
				}

				if err := store.DeleteModelActivationStatusInTransaction(tx, k); err != nil {
					// Gracefully handle a not-found error for backward compatibility.
					if !errors.Is(err, gorm.ErrRecordNotFound) {
						return status.Errorf(codes.NotFound, "model activation status for %q not found", req.Id)
					}
				}

				// TODO(eknji): Delete the model config.

				return nil
			}); err != nil {
				return nil, err
			}
		}
	case *v1.UpdateBaseModelLoadingStatusRequest_Failure_:
		failure := req.GetFailure()
		if failure.Reason == "" {
			return nil, status.Error(codes.InvalidArgument, "reason is required")
		}
		err = s.store.UpdateBaseModelToFailedStatus(k, failure.Reason)
	default:
		// Loading is still in progress. Just update the status message.
		err = s.store.UpdateBaseModelLoadingStatusMessage(k, req.StatusMessage)
	}

	if err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update base model loading status: %s", err)
	}

	return &v1.UpdateBaseModelLoadingStatusResponse{}, nil
}

// UpdateModelLoadingStatus updates the loading status. When the loading succeeded, it also
// updates the model metadata.
func (s *WS) UpdateModelLoadingStatus(
	ctx context.Context,
	req *v1.UpdateModelLoadingStatusRequest,
) (*v1.UpdateModelLoadingStatusResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.LoadingResult == nil && req.StatusMessage == "" {
		return nil, status.Error(codes.InvalidArgument, "loading_result or status_message is required")
	}

	if _, err := s.store.GetModelByModelIDAndTenantID(req.Id, clusterInfo.TenantID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	switch req.LoadingResult.(type) {
	case *v1.UpdateModelLoadingStatusRequest_Success_:
		err = s.store.UpdateModelToSucceededStatus(
			req.Id,
			clusterInfo.TenantID,
		)
	case *v1.UpdateModelLoadingStatusRequest_Failure_:
		failure := req.GetFailure()
		if failure.Reason == "" {
			return nil, status.Error(codes.InvalidArgument, "reason is required")
		}
		err = s.store.UpdateModelToFailedStatus(
			req.Id,
			clusterInfo.TenantID,
			failure.Reason,
		)
	default:
		// Loading is still in progress. Just update the status message.
		err = s.store.UpdateModelLoadingStatusMessage(req.Id, clusterInfo.TenantID, req.StatusMessage)
	}

	if err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update model loading status: %s", err)
	}

	return &v1.UpdateModelLoadingStatusResponse{}, nil
}

func (s *WS) getLoadedBaseModel(modelID, tenantID string) (*store.BaseModel, bool, error) {
	// Find all base models regardless of the project.
	bms, err := s.store.ListBaseModelsByModelIDAndTenantID(modelID, tenantID)
	if err != nil {
		return nil, false, fmt.Errorf("list base models by model ID and tenant ID: %w", err)
	}

	for _, bm := range bms {
		if isBaseModelLoaded(bm) {
			return bm, true, nil
		}
	}

	return nil, false, nil
}

func (s *WS) generateModelID(baseModel, suffix, tenantID string) (string, error) {
	const randomLength = 10
	// OpenAI uses ':" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	base := fmt.Sprintf("ft:%s:%s-", mid.ToLLMarinerModelID(baseModel), suffix)

	// Randomly create an ID and retry if it already exists.
	for {
		randomStr, err := id.GenerateID("", randomLength)
		if err != nil {
			return "", fmt.Errorf("generate ID: %s", err)
		}
		id := fmt.Sprintf("%s%s", base, randomStr)
		if _, err := s.store.GetModelByModelIDAndTenantID(id, tenantID); errors.Is(err, gorm.ErrRecordNotFound) {
			return id, nil
		}
	}
}

// GetModel gets a model.
func (s *WS) GetModel(ctx context.Context, req *v1.GetModelRequest) (*v1.Model, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	bm, found, err := s.getLoadedBaseModel(req.Id, clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get loaded base model: %s", err)
	}
	if found {
		return convertBaseModelToProto(s.store, s.pcache, bm)
	}

	// Try a fine-tuned model next.
	fm, err := s.store.GetPublishedModelByModelIDAndTenantID(req.Id, clusterInfo.TenantID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get model by model ID: %s", err)
		}
		return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
	}

	if !isModelLoaded(fm) {
		return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
	}

	return convertFineTunedModelToProto(s.store, s.pcache, fm)
}

func toBaseModelProto(m *store.BaseModel) *v1.BaseModel {
	return &v1.BaseModel{
		Id:      m.ModelID,
		Object:  "basemodel",
		Created: m.CreatedAt.UTC().Unix(),
	}
}
