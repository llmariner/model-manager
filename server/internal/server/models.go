package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/common/pkg/id"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

// CreateModel creates a base model.
func (s *S) CreateModel(
	ctx context.Context,
	req *v1.CreateModelRequest,
) (*v1.Model, error) {
	if req.IsFineTunedModel {
		return s.createFineTunedModel(ctx, req)

	}
	return s.createBaseModel(ctx, req)
}

func (s *S) createFineTunedModel(
	ctx context.Context,
	req *v1.CreateModelRequest,
) (*v1.Model, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.BaseModelId == "" {
		return nil, status.Error(codes.InvalidArgument, "base_model_id is required")
	}

	if req.Suffix == "" {
		return nil, status.Error(codes.InvalidArgument, "suffix is required")
	}
	// TODO(kenji): This follows the OpenAI API spec, but might not be necessary.
	// TODO(kenji): Longer suffix needs to be truncated when a statefulset name is generated. Make sure there
	// is no conflict.
	if len(req.Suffix) > 18 {
		return nil, status.Errorf(codes.InvalidArgument, "suffix must not be more than 18 characters")
	}

	// Only support S3 as a source repository for now.
	if req.SourceRepository != v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported source repository: %s", req.SourceRepository)
	}

	// TODO(kenji): Revisit.
	if req.ModelFileLocation == "" {
		return nil, status.Error(codes.InvalidArgument, "model_file_location is required")
	}
	if !strings.HasPrefix(req.ModelFileLocation, "s3://") {
		return nil, status.Errorf(codes.InvalidArgument, "model file location must start with s3://, but got %s", req.ModelFileLocation)
	}

	if err := validateModelConfig(req.Config); err != nil {
		return nil, err
	}

	if _, found, err := getVisibleBaseModel(s.store, req.BaseModelId, userInfo, true /* includeLoadingModel */); err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	} else if !found {
		return nil, status.Errorf(codes.NotFound, "base model %q not found", req.BaseModelId)
	}

	id := fmt.Sprintf("ft:%s:%s", req.BaseModelId, req.Suffix)

	sc, err := s.store.GetStorageConfig(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get storage config: %s", err)
	}

	path := fmt.Sprintf("%s/%s/%s/%s", sc.PathPrefix, userInfo.TenantID, userInfo.ProjectID, id)

	var m *store.Model
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		var err error
		m, err = store.CreateModelInTransaction(tx, store.ModelSpec{
			ModelID:           id,
			TenantID:          userInfo.TenantID,
			OrganizationID:    userInfo.OrganizationID,
			ProjectID:         userInfo.ProjectID,
			IsPublished:       true,
			Path:              path,
			BaseModelID:       req.BaseModelId,
			Adapter:           v1.AdapterType_ADAPTER_TYPE_LORA,
			LoadingStatus:     v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
			SourceRepository:  req.SourceRepository,
			ModelFileLocation: req.ModelFileLocation,
		})
		if err != nil {
			if gerrors.IsUniqueConstraintViolation(err) {
				return status.Errorf(codes.AlreadyExists, "model %q already exists", id)
			}
			return status.Errorf(codes.Internal, "create model: %s", err)
		}

		if err := store.CreateModelActivationStatusInTransaction(tx, &store.ModelActivationStatus{
			ModelID: id,
			// ProjectID is empty for fine-tuned models for backward compatibility.
			TenantID: userInfo.TenantID,
			Status:   v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
		}); err != nil {
			return status.Errorf(codes.Internal, "create model activation status: %s", err)
		}

		if c := req.Config; c != nil {
			if err := createModelConfigInTransaction(tx, store.ModelKey{
				ModelID: id,
				// ProjectID is empty for fine-tuned models.
				TenantID: userInfo.TenantID,
			}, c); err != nil {
				return status.Errorf(codes.Internal, "create model config: %s", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	proj, err := s.pcache.GetProject(userInfo.ProjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get project: %s", err)
	}

	return toModelProto(m, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, req.Config, proj), nil
}

func (s *S) createBaseModel(
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

	if err := validateIDAndSourceRepository(req.Id, req.SourceRepository); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	if err := validateModelConfig(req.Config); err != nil {
		return nil, err
	}

	var projectID string
	if req.IsProjectScoped {
		projectID = userInfo.ProjectID
	}

	// Check if the model already exists (with the original name as we as the converted name).
	for _, modelID := range []string{req.Id, id.ToLLMarinerModelID(req.Id)} {
		k := store.ModelKey{
			ModelID:   modelID,
			ProjectID: projectID,
			TenantID:  userInfo.TenantID,
		}
		if _, err := s.store.GetBaseModel(k); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, status.Errorf(codes.Internal, "get base model: %s", err)
			}
		} else {
			return nil, status.Errorf(codes.AlreadyExists, "base model %q already exists", req.Id)
		}
	}

	var m *store.BaseModel
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		k := store.ModelKey{
			ModelID:   req.Id,
			ProjectID: projectID,
			TenantID:  userInfo.TenantID,
		}
		var err error
		m, err = store.CreateBaseModelWithLoadingRequestedInTransaction(tx, k, req.SourceRepository)
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

		if c := req.Config; c != nil {
			if err := createModelConfigInTransaction(tx, k, c); err != nil {
				return status.Errorf(codes.Internal, "create model config: %s", err)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var proj *v1.Project
	if projectID != "" {
		var err error
		proj, err = s.pcache.GetProject(projectID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get project: %s", err)
		}
	}

	mp, err := baseToModelProto(m, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, req.Config, proj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "base model to proto: %s", err)
	}
	return mp, nil
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

	if req.Limit < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "limit must be non-negative")
	}
	limit := req.Limit
	if limit == 0 {
		limit = defaultPageSize
	}
	if limit > maxPageSize {
		limit = maxPageSize
	}

	// Validate the "after" parameter if provided.
	var (
		afterBase  *store.BaseModel
		afterModel *store.Model
	)
	if req.After != "" {
		var found bool
		var err error
		afterBase, afterModel, found, err = getVisibleModel(s.store, req.After, userInfo, req.IncludeLoadingModels)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "find model by ID %q: %s", req.After, err)
		}
		if !found {
			return nil, status.Errorf(codes.InvalidArgument, "after parameter %q is not a valid model ID", req.After)
		}
	}

	totalItems, err := s.getTotalItems(userInfo, req.IncludeLoadingModels)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get total items: %s", err)
	}

	// Categories for sorting models
	categories := []struct {
		isBase bool
		status v1.ActivationStatus
	}{
		{true, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE},
		{false, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE},
		{true, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE},
		{false, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE},
	}

	// Returns the index of the category based on whether it's a base model and its activation status.
	var activationCategory = func(isBase bool, status v1.ActivationStatus) int {
		for i, cat := range categories {
			if cat.isBase == isBase && cat.status == status {
				return i
			}
		}
		return -1 // or return error
	}

	// Declare output variables before transaction scope
	var (
		modelProtos []*v1.Model
		hasMore     bool
	)

	// Determine what models to list in a transaction in case models change state
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		// Starting category for sorting
		var (
			startCat int
			afterID  string
		)
		if req.After != "" {
			var k store.ModelKey
			if afterBase != nil {
				k = store.ModelKey{
					ModelID:   afterBase.ModelID,
					ProjectID: afterBase.ProjectID,
					TenantID:  afterBase.TenantID,
				}
			} else {
				k = store.ModelKey{
					ModelID: afterModel.ModelID,
					// ProjectID is empty for fine-tuned models (for backward compatibility).
					TenantID: afterModel.TenantID,
				}
			}

			as, err := getModelActivationStatusInTransaction(tx, k)
			if err != nil {
				return status.Errorf(codes.Internal, "get model activation status: %s", err)
			}
			startCat = activationCategory(afterBase != nil, as)
			afterID = req.After
		}

		for i := startCat; i < len(categories); i++ {
			// TODO: Investigate a possible speed up by using orderby activation status on both queries
			// up to the max number, then layering the results
			cat := categories[i]
			mps, more, err := listModelsByActivationStatus(
				tx,
				s.pcache,
				userInfo.ProjectID,
				userInfo.TenantID,
				req.IncludeLoadingModels,
				cat.isBase,
				cat.status,
				afterID,
				int(limit)-len(modelProtos),
			)
			if err != nil {
				return status.Errorf(codes.Internal, "list models: %s", err)
			}
			modelProtos = append(modelProtos, mps...)

			if more {
				hasMore = true
				break
			}

			if len(modelProtos) == int(limit) {
				// No need to query further but we don't know if hasMore is true or false yet.
				// See if there are more models in future categories.
				for j := i; j < len(categories); j++ {
					cat := categories[j]
					mps, _, err := listModelsByActivationStatus(
						tx,
						s.pcache,
						userInfo.ProjectID,
						userInfo.TenantID,
						req.IncludeLoadingModels,
						cat.isBase,
						cat.status,
						"", /* afterID */
						1,  /* limit */
					)
					if err != nil {
						return status.Errorf(codes.Internal, "list models: %s", err)
					}
					if len(mps) > 0 {
						hasMore = true
						break
					}
				}

				break
			}

			// Reset afterID for the next category.
			afterID = ""
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &v1.ListModelsResponse{
		Object:     "list",
		Data:       modelProtos,
		HasMore:    hasMore,
		TotalItems: totalItems,
	}, nil
}

func (s *S) getTotalItems(userInfo *auth.UserInfo, includeLoadingModels bool) (int32, error) {
	totalModels, err := s.store.CountModelsByProjectID(userInfo.ProjectID, true, includeLoadingModels)
	if err != nil {
		return 0, err
	}

	totalBaseModels, err := s.store.CountBaseModels(userInfo.ProjectID, userInfo.TenantID, includeLoadingModels)
	if err != nil {
		return 0, err
	}
	return int32(totalModels + totalBaseModels), nil
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

	m, err := getVisibleModelProto(s.store, s.pcache, req.Id, userInfo, req.IncludeLoadingModel)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// UpdateModel updates a model.
func (s *S) UpdateModel(
	ctx context.Context,
	req *v1.UpdateModelRequest,
) (*v1.Model, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Model == nil {
		return nil, status.Error(codes.InvalidArgument, "model is required")
	}

	if req.Model.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := validateModelConfig(req.Model.Config); err != nil {
		return nil, err
	}

	// Currently only support the update of the config field.
	if req.UpdateMask == nil {
		return nil, status.Error(codes.InvalidArgument, "update mask is required")
	}

	// Only allow updating the config field and the config must be present.
	if req.Model.Config == nil {
		return nil, status.Error(codes.InvalidArgument, "model config is required")
	}

	k, err := getKeyForVisibleModel(s.store, req.Model.Id, userInfo, true /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}

	var (
		hasExisting bool
		config      v1.ModelConfig
	)
	if c, err := s.store.GetModelConfig(k); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get model config: %s", err)
		}
		config = *defaultModelConfig()
	} else {
		hasExisting = true
		if err := proto.Unmarshal(c.EncodedConfig, &config); err != nil {
			return nil, err
		}
	}

	patchedConfig, err := patchModelConfig(&config, req.Model.Config, req.UpdateMask)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err)
	}

	b, err := proto.Marshal(patchedConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "marshal model config: %s", err)
	}

	if hasExisting {
		if err := s.store.UpdateModelConfig(k, b); err != nil {
			return nil, status.Errorf(codes.Internal, "update model config: %s", err)
		}
	} else {
		if err := s.store.CreateModelConfig(&store.ModelConfig{
			ModelID:       k.ModelID,
			ProjectID:     k.ProjectID,
			TenantID:      k.TenantID,
			EncodedConfig: b,
		}); err != nil {
			return nil, status.Errorf(codes.Internal, "create model config: %s", err)
		}
	}

	m, err := getVisibleModelProto(s.store, s.pcache, req.Model.Id, userInfo, true /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}
	return m, nil
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

	k, err := getKeyForVisibleModel(s.store, req.Id, userInfo, true /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}

	if as, err := getModelActivationStatus(s.store, k); err != nil {
		return nil, status.Errorf(codes.Internal, "get model activation status: %s", err)
	} else if as == v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE {
		return nil, status.Errorf(codes.FailedPrecondition, "model %q is active", req.Id)
	}

	if _, err := s.store.GetBaseModel(k); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}

		// The specified model is not a base-model or the base-model has already been deleted.
		// Try deleting a fine-tuned model of the specified ID.
		return s.deleteFineTunedModel(ctx, k)
	}

	// The specified model is a base-model. Delete it.

	return s.deleteBaseModel(ctx, k)
}

func (s *S) deleteFineTunedModel(ctx context.Context, k store.ModelKey) (*v1.DeleteModelResponse, error) {
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteModelInTransaction(tx, k.ModelID, k.TenantID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.NotFound, "model %q not found", k.ModelID)
			}
			return status.Errorf(codes.Internal, "delete model: %s", err)
		}
		if err := store.DeleteModelActivationStatusInTransaction(tx, k); err != nil {
			// Gracefully handle a not found error for backward compatibility.
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.NotFound, "model activation status %q not found", k.ModelID)
			}
		}

		// TODO(kenji): Delete model config

		return nil
	}); err != nil {
		return nil, err
	}

	return &v1.DeleteModelResponse{
		Id:      k.ModelID,
		Object:  "model",
		Deleted: true,
	}, nil
}

func (s *S) deleteBaseModel(ctx context.Context, k store.ModelKey) (*v1.DeleteModelResponse, error) {
	// TODO(kenji): Revisit the permission check. The base model is scoped by a tenant, not project,
	// so we should have additional check here.
	if err := s.store.Transaction(func(tx *gorm.DB) error {
		if err := store.DeleteBaseModelInTransaction(tx, k); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.Internal, "delete model: %s", err)
			}
			return status.Errorf(codes.NotFound, "model %q not found", k.ModelID)
		}

		// Delete the HFModelRepo if the model is from Hugging Face. Otherwise the same
		// model cannot be reloaded again.
		//
		// TODO(kenji): Handle a case where a single Hugging Face repo has multiple models. In that case,
		// the Hugging Face repo name and the model ID does not match.
		//
		// Also, deleting a HFModelRepo can trigger downloading the remaining undeleted models again, which is not ideal.
		if err := store.DeleteHFModelRepoInTransactionByModelID(tx, k); err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.Internal, "delete hf model repo (id: %q): %s", k.ModelID, err)
			}
			// Ignore. The HFModelRepo does not exist for old models or non-HF models.
		}

		if err := store.DeleteModelActivationStatusInTransaction(tx, k); err != nil {
			// Gracefully handle a not found error for backward compatibility.
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return status.Errorf(codes.NotFound, "model activation status for %q not found", k.ModelID)
			}
		}

		// TODO(kenji): Delete model config.

		return nil
	}); err != nil {
		return nil, err
	}

	return &v1.DeleteModelResponse{
		Id:      k.ModelID,
		Object:  "model",
		Deleted: true,
	}, nil
}

// ActivateModel activates a model.
func (s *S) ActivateModel(ctx context.Context, req *v1.ActivateModelRequest) (*v1.ActivateModelResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	k, err := getKeyForVisibleModel(s.store, req.Id, userInfo, false /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}

	if err := s.store.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE); err != nil {
		return nil, status.Errorf(codes.Internal, "update model activation status: %s", err)
	}

	return &v1.ActivateModelResponse{}, nil
}

// DeactivateModel deactivates a model.
func (s *S) DeactivateModel(ctx context.Context, req *v1.DeactivateModelRequest) (*v1.DeactivateModelResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	k, err := getKeyForVisibleModel(s.store, req.Id, userInfo, false /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}

	if err := s.store.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE); err != nil {
		return nil, status.Errorf(codes.Internal, "update model activation status: %s", err)
	}

	return &v1.DeactivateModelResponse{}, nil
}

// listModelsByActivationStatus lists models by activation status with pagination.
// This is a helper function called by ListModels.
func listModelsByActivationStatus(
	tx *gorm.DB,
	pcache pcache,
	projectID string,
	tenantID string,
	includeLoadingModels bool,
	isBaseModel bool,
	activationStatus v1.ActivationStatus,
	afterID string,
	limit int,
) ([]*v1.Model, bool, error) {
	if isBaseModel {
		bms, more, err := store.ListBaseModelsByActivationStatusWithPaginationInTransaction(
			tx,
			projectID,
			tenantID,
			activationStatus,
			afterID,
			limit,
			includeLoadingModels,
		)
		if err != nil {
			return nil, false, fmt.Errorf("list base models: %s", err)
		}

		var modelProtos []*v1.Model
		for _, m := range bms {
			mc, err := getModelConfigInTransaction(tx, store.ModelKey{
				ModelID:   m.ModelID,
				ProjectID: m.ProjectID,
				TenantID:  m.TenantID,
			})
			if err != nil {
				return nil, false, fmt.Errorf("get model config: %s", err)
			}

			var proj *v1.Project
			if m.ProjectID != "" {
				proj, err = pcache.GetProject(m.ProjectID)
				if err != nil {
					return nil, false, fmt.Errorf("get project: %s", err)
				}
			}

			mp, err := baseToModelProto(m, activationStatus, mc, proj)
			if err != nil {
				return nil, false, fmt.Errorf("to proto: %s", err)
			}
			modelProtos = append(modelProtos, mp)
		}
		return modelProtos, more, nil
	}

	ms, more, err := store.ListModelsByActivationStatusWithPaginationInTransaction(
		tx,
		projectID,
		true, /* onlyPublished */
		activationStatus,
		afterID,
		limit,
		includeLoadingModels,
	)
	if err != nil {
		return nil, false, fmt.Errorf("list models: %s", err)
	}

	var modelProtos []*v1.Model
	for _, m := range ms {
		mc, err := getModelConfigInTransaction(tx, store.ModelKey{
			ModelID: m.ModelID,
			// ProjectID is empty for fine-tuned models.
			TenantID: m.TenantID,
		})
		if err != nil {
			return nil, false, fmt.Errorf("get model config: %s", err)
		}

		proj, err := pcache.GetProject(m.ProjectID)
		if err != nil {
			return nil, false, fmt.Errorf("get project: %s", err)
		}

		modelProtos = append(modelProtos, toModelProto(m, activationStatus, mc, proj))
	}

	return modelProtos, more, nil
}

func getKeyForVisibleModel(
	st *store.S,
	modelID string,
	userInfo *auth.UserInfo,
	includeLoadingModel bool,
) (store.ModelKey, error) {
	bm, fm, found, err := getVisibleModel(st, modelID, userInfo, includeLoadingModel)
	if err != nil {
		return store.ModelKey{}, status.Errorf(codes.Internal, "%s", err)
	}
	if !found {
		return store.ModelKey{}, status.Errorf(codes.NotFound, "model %q not found", modelID)
	}

	if bm != nil {
		return store.ModelKey{
			ModelID:   bm.ModelID,
			ProjectID: bm.ProjectID,
			TenantID:  bm.TenantID,
		}, nil
	}

	return store.ModelKey{
		ModelID: fm.ModelID,
		// ProjectID is empty for backward compatibility.
		TenantID: fm.TenantID,
	}, nil
}

func getVisibleModelProto(
	st *store.S,
	pcache pcache,
	modelID string,
	userInfo *auth.UserInfo,
	includeLoadingModel bool,
) (*v1.Model, error) {
	bm, fm, found, err := getVisibleModel(st, modelID, userInfo, includeLoadingModel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%s", err)
	}
	if !found {
		return nil, status.Errorf(codes.NotFound, "model %q not found", modelID)
	}

	if bm != nil {
		return convertBaseModelToProto(st, pcache, bm)
	}

	return convertFineTunedModelToProto(st, pcache, fm)
}

func convertBaseModelToProto(
	st *store.S,
	pcache pcache,
	m *store.BaseModel,
) (*v1.Model, error) {
	var projectID string
	if m.ProjectID != "" {
		projectID = m.ProjectID
	}

	k := store.ModelKey{
		ModelID:   m.ModelID,
		ProjectID: projectID,
		TenantID:  m.TenantID,
	}
	as, err := getModelActivationStatus(st, k)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get model activation status: %s", err)
	}

	mc, err := getModelConfig(st, k)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get model config: %s", err)
	}

	var proj *v1.Project
	if projectID != "" {
		proj, err = pcache.GetProject(projectID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get project: %s", err)
		}
	}

	mp, err := baseToModelProto(m, as, mc, proj)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "to proto: %s", err)
	}
	return mp, nil
}

func convertFineTunedModelToProto(
	st *store.S,
	pcache pcache,
	m *store.Model,
) (*v1.Model, error) {
	k := store.ModelKey{
		ModelID: m.ModelID,
		// ProjectID is empty for fine-tuned models for backward compatibility.
		TenantID: m.TenantID,
	}
	as, err := getModelActivationStatus(st, k)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get model activation status: %s", err)
	}

	// TODO(kenji): Consider using the model config of the base model if the fine-tuned
	// model doesn't have the model config.

	mc, err := getModelConfig(st, k)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get model config: %s", err)
	}

	proj, err := pcache.GetProject(m.ProjectID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get project: %s", err)
	}

	return toModelProto(m, as, mc, proj), nil
}

func getModelActivationStatus(st *store.S, k store.ModelKey) (v1.ActivationStatus, error) {
	status, err := st.GetModelActivationStatus(k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.ActivationStatus_ACTIVATION_STATUS_UNSPECIFIED, err
		}
		// For backward compatibility.
		return v1.ActivationStatus_ACTIVATION_STATUS_UNSPECIFIED, nil
	}
	return status.Status, nil
}

func getModelActivationStatusInTransaction(tx *gorm.DB, k store.ModelKey) (v1.ActivationStatus, error) {
	status, err := store.GetModelActivationStatusInTransaction(tx, k)
	if err != nil {
		return v1.ActivationStatus_ACTIVATION_STATUS_UNSPECIFIED, err
	}
	return status.Status, nil
}

func getModelConfig(st *store.S, k store.ModelKey) (*v1.ModelConfig, error) {
	c, err := st.GetModelConfig(k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return defaultModelConfig(), nil
	}

	var config v1.ModelConfig
	if err := proto.Unmarshal(c.EncodedConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getModelConfigInTransaction(tx *gorm.DB, k store.ModelKey) (*v1.ModelConfig, error) {
	c, err := store.GetModelConfigInTransaction(tx, k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return defaultModelConfig(), nil
	}

	var config v1.ModelConfig
	if err := proto.Unmarshal(c.EncodedConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func defaultModelConfig() *v1.ModelConfig {
	return &v1.ModelConfig{
		RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
			Resources: &v1.ModelConfig_RuntimeConfig_Resources{
				Gpu: 1,
			},
			Replicas: 1,
		},
		ClusterAllocationPolicy: &v1.ModelConfig_ClusterAllocationPolicy{
			EnableOnDemandAllocation: true,
		},
	}
}

func createModelConfigInTransaction(tx *gorm.DB, k store.ModelKey, c *v1.ModelConfig) error {
	b, err := proto.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal model config: %s", err)
	}
	if err := store.CreateModelConfigInTransaction(tx, &store.ModelConfig{
		ModelID:       k.ModelID,
		ProjectID:     k.ProjectID,
		TenantID:      k.TenantID,
		EncodedConfig: b,
	}); err != nil {
		return fmt.Errorf("create model config: %s", err)
	}
	return nil
}

func getVisibleModel(
	st *store.S,
	modelID string,
	userInfo *auth.UserInfo,
	includeLoadingModels bool,
) (*store.BaseModel, *store.Model, bool, error) {
	bm, found, err := getVisibleBaseModel(st, modelID, userInfo, includeLoadingModels)
	if err != nil {
		return nil, nil, false, err
	}
	if found {
		return bm, nil, true, nil
	}

	// Try a fine-tuned model next.
	m, err := st.GetPublishedModelByModelIDAndTenantID(modelID, userInfo.TenantID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, false, err
		}

		return nil, nil, false, nil
	}

	if !isModelLoaded(m) && !includeLoadingModels {
		return nil, nil, false, nil
	}

	return nil, m, true, nil
}

func getVisibleBaseModel(
	st *store.S,
	modelID string,
	userInfo *auth.UserInfo,
	includeLoadingModel bool,
) (*store.BaseModel, bool, error) {
	k := store.ModelKey{
		ModelID:   modelID,
		ProjectID: userInfo.ProjectID,
		TenantID:  userInfo.TenantID,
	}
	if bm, err := st.GetBaseModel(k); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, fmt.Errorf("get base model: %s", err)
		}
	} else if isBaseModelLoaded(bm) || includeLoadingModel {
		return bm, true, nil
	}

	// Check the global-scoped model.
	k.ProjectID = ""
	bm, err := st.GetBaseModel(k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, fmt.Errorf("get base model: %s", err)
		}
		return nil, false, nil
	}

	if isBaseModelLoaded(bm) || includeLoadingModel {
		return bm, true, nil
	}

	return nil, false, nil
}

func toModelProto(m *store.Model, as v1.ActivationStatus, config *v1.ModelConfig, project *v1.Project) *v1.Model {
	var statusMsg string
	if m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING || m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED {
		statusMsg = m.LoadingStatusMessage
	}

	return &v1.Model{
		Id:                   m.ModelID,
		Object:               "model",
		Created:              m.CreatedAt.UTC().Unix(),
		OwnedBy:              "user",
		LoadingStatus:        toLoadingStatus(m.LoadingStatus),
		SourceRepository:     v1.SourceRepository_SOURCE_REPOSITORY_FINE_TUNING,
		LoadingFailureReason: m.LoadingFailureReason,
		LoadingStatusMessage: statusMsg,
		// Fine-tuned models always have the Hugging Face format.
		Formats:     []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		IsBaseModel: false,
		BaseModelId: m.BaseModelID,

		ActivationStatus: as,

		Config: config,

		Project: project,
	}
}

func baseToModelProto(m *store.BaseModel, as v1.ActivationStatus, config *v1.ModelConfig, project *v1.Project) (*v1.Model, error) {
	formats, err := store.UnmarshalModelFormats(m.Formats)
	if err != nil {
		return nil, err
	}
	if len(formats) == 0 {
		// For backward compatibility.
		formats = []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}
	}

	var statusMsg string
	if m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING || m.LoadingStatus == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED {
		statusMsg = m.LoadingStatusMessage
	}

	return &v1.Model{
		Id:                   m.ModelID,
		Object:               "model",
		Created:              m.CreatedAt.UTC().Unix(),
		OwnedBy:              "system",
		LoadingStatus:        toLoadingStatus(m.LoadingStatus),
		SourceRepository:     m.SourceRepository,
		LoadingFailureReason: m.LoadingFailureReason,
		LoadingStatusMessage: statusMsg,
		Formats:              formats,
		IsBaseModel:          true,
		BaseModelId:          "",

		ActivationStatus: as,

		Config: config,

		Project: project,
	}, nil
}

func toLoadingStatus(s v1.ModelLoadingStatus) v1.ModelLoadingStatus {
	if s == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_UNSPECIFIED {
		// The UNSPECIFIED status is considered as loaded for backward compatibility.
		return v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED
	}
	return s
}

func isBaseModelLoaded(m *store.BaseModel) bool {
	return toLoadingStatus(m.LoadingStatus) == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED
}

func isModelLoaded(m *store.Model) bool {
	return toLoadingStatus(m.LoadingStatus) == v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED
}

func validateIDAndSourceRepository(id string, sourceRepository v1.SourceRepository) error {
	switch sourceRepository {
	case v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE:
		if strings.HasPrefix("s3://", id) {
			// TODO(kenji): This is not very intuitive. Model manager loader instead should be able to
			// download from any bucket that a user specifies.
			return fmt.Errorf("id must not include s3://<bucket>")
		}
	case v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE:
		l := strings.Split(id, "/")
		if len(l) <= 1 || len(l) > 3 {
			return fmt.Errorf("unexpected model ID format: %s. The format should be <org>/<repo> or <org>/<repo>/<file>", id)
		}
		if l[0] == "" || l[1] == "" {
			return fmt.Errorf("unexpected model ID format: %s. The format should be <org>/<repo> or <org>/<repo>/<file>", id)
		}
	case v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA:
		l := strings.Split(id, ":")
		if len(l) != 2 {
			return fmt.Errorf("unexpected model ID format: %s. The format should be <model>:<tag>", id)
		}
		if l[0] == "" || l[1] == "" {
			return fmt.Errorf("unexpected model ID format: %s. The format should be <model>:<tag>", id)
		}
	default:
		return fmt.Errorf("source_repository must be one of %v", []v1.SourceRepository{
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA,
		})
	}
	return nil
}

func validateModelConfig(c *v1.ModelConfig) error {
	if c == nil {
		// Do nothing. ModelConfig is optional.
		return nil
	}

	rc := c.RuntimeConfig
	if rc == nil {
		return status.Error(codes.InvalidArgument, "runtime config is required")
	}

	if r := rc.Resources; r != nil {
		if r.Gpu < 0 {
			return status.Error(codes.InvalidArgument, "gpu must be greater than or equal to 0")
		}
	}
	if rc.Replicas <= 0 {
		return status.Error(codes.InvalidArgument, "replicas must be greater than 0")
	}

	cap := c.ClusterAllocationPolicy
	if cap == nil {
		return status.Error(codes.InvalidArgument, "cluster allocation policy is required")
	}

	return nil
}

func patchModelConfig(
	config *v1.ModelConfig,
	patch *v1.ModelConfig,
	updateMask *fieldmaskpb.FieldMask,
) (*v1.ModelConfig, error) {
	if config.RuntimeConfig == nil {
		config.RuntimeConfig = &v1.ModelConfig_RuntimeConfig{}
	}
	if config.RuntimeConfig.Resources == nil {
		config.RuntimeConfig.Resources = &v1.ModelConfig_RuntimeConfig_Resources{}
	}
	if config.ClusterAllocationPolicy == nil {
		config.ClusterAllocationPolicy = &v1.ModelConfig_ClusterAllocationPolicy{}
	}

	for _, path := range updateMask.Paths {
		switch {
		case strings.HasPrefix(path, "config"):
			if patch == nil {
				return nil, fmt.Errorf("config is required")
			}

			if path == "config" {
				config = patch
				break
			}

			cpath := path[len("config."):]
			switch {
			case strings.HasPrefix(cpath, "runtime_config"):
				rc := patch.RuntimeConfig
				if rc == nil {
					return nil, fmt.Errorf("runtime_config is required")
				}

				if cpath == "runtime_config" {
					config.RuntimeConfig = rc
					break
				}

				rpath := cpath[len("runtime_config."):]
				switch {
				case rpath == "replicas":
					r := rc.Replicas
					if r <= 0 {
						return nil, fmt.Errorf("runtime_config.replicas must be positive, but got %d", r)
					}
					config.RuntimeConfig.Replicas = r
				case strings.HasPrefix(rpath, "resources"):
					r := rc.Resources
					if r == nil {
						return nil, fmt.Errorf("runtime_config.resources is required")
					}

					if rpath == "resources" {
						config.RuntimeConfig.Resources = rc.Resources
						break
					}

					rrpath := rpath[len("resources."):]
					switch rrpath {
					case "gpu":
						v := r.Gpu
						if v < 0 {
							return nil, fmt.Errorf("runtime_config.resources.gpu must be non-negative, but got %d", v)
						}
						config.RuntimeConfig.Resources.Gpu = v
					default:
						return nil, fmt.Errorf("unsupported update mask path: %s", path)
					}
					// TODO(kenji): support extra_args
				default:
					return nil, fmt.Errorf("unsupported update mask path: %s", path)
				}
			case strings.HasPrefix(cpath, "cluster_allocation_policy"):
				rc := patch.ClusterAllocationPolicy
				if rc == nil {
					return nil, fmt.Errorf("cluster_allocation_policy is required")
				}

				if cpath == "cluster_allocation_policy" {
					config.ClusterAllocationPolicy = rc
					break
				}

				ppath := cpath[len("cluster_allocation_policy."):]
				switch ppath {
				case "enable_on_demand_allocation":
					config.ClusterAllocationPolicy.EnableOnDemandAllocation = rc.EnableOnDemandAllocation
					// TODO(kenji): Support allowed_cluster_ids
				default:
					return nil, fmt.Errorf("unsupported update mask path: %s", path)
				}
			default:
				return nil, fmt.Errorf("unsupported update mask path: %s", path)
			}
		default:
			return nil, fmt.Errorf("unsupported update mask path: %s", path)
		}
	}

	d := defaultModelConfig()

	// Set default values to nil fields.
	if config.RuntimeConfig == nil {
		config.RuntimeConfig = d.RuntimeConfig
	}
	if config.ClusterAllocationPolicy == nil {
		config.ClusterAllocationPolicy = d.ClusterAllocationPolicy
	}

	return config, nil
}
