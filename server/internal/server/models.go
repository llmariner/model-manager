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

	k := store.ModelKey{
		ModelID:  req.BaseModelId,
		TenantID: userInfo.TenantID,
	}
	if _, err := s.store.GetBaseModel(k); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}
		return nil, status.Errorf(codes.NotFound, "base model %q not found", req.BaseModelId)
	}

	id := fmt.Sprintf("ft:%s:%s", req.BaseModelId, req.Suffix)

	sc, err := s.store.GetStorageConfig(userInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get storage config: %s", err)
	}

	path := fmt.Sprintf("%s/%s/%s", sc.PathPrefix, userInfo.TenantID, id)

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
			ModelID:  id,
			TenantID: userInfo.TenantID,
			Status:   v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
		}); err != nil {
			return status.Errorf(codes.Internal, "create model activation status: %s", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return toModelProto(m, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE), nil
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

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: userInfo.TenantID,
	}
	if _, err := s.store.GetBaseModel(k); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}
	} else {
		return nil, status.Errorf(codes.AlreadyExists, "base model %q already exists", req.Id)
	}

	var m *store.BaseModel
	if err := s.store.Transaction(func(tx *gorm.DB) error {
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

		return nil
	}); err != nil {
		return nil, err
	}

	mp, err := baseToModelProto(m, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE)
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
		var err error
		afterBase, afterModel, err = s.findBaseModelOrModel(req.After, userInfo.TenantID, req.IncludeLoadingModels)
		if err != nil {
			return nil, err
		}
	}

	totalItems, err := s.getTotalItems(userInfo)
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
			if afterBase != nil {
				k := store.ModelKey{
					ModelID:  afterBase.ModelID,
					TenantID: afterBase.TenantID,
				}
				as, err := getModelActivationStatusInTransaction(tx, k)
				if err != nil {
					return status.Errorf(codes.Internal, "get model activation status: %s", err)
				}
				startCat = activationCategory(true, as)
			} else {
				k := store.ModelKey{
					ModelID:  afterModel.ModelID,
					TenantID: afterModel.TenantID,
				}
				as, err := getModelActivationStatusInTransaction(tx, k)
				if err != nil {
					return status.Errorf(codes.Internal, "get model activation status: %s", err)
				}
				startCat = activationCategory(false, as)
			}
			afterID = req.After
		}

		endCat := -1
		for i := startCat; i < len(categories); i++ {
			// TODO: Investigate a possible speed up by using orderby activation status on both queries up to the max number, then layering the results
			var more bool
			if cat := categories[i]; cat.isBase {
				var bms []*store.BaseModel
				bms, more, err = store.ListBaseModelsByActivationStatusWithPaginationInTransaction(
					tx,
					userInfo.TenantID,
					cat.status,
					afterID,
					int(limit)-len(modelProtos),
					req.IncludeLoadingModels,
				)
				if err != nil {
					return status.Errorf(codes.Internal, "list base models: %s", err)
				}
				for _, m := range bms {
					mp, err := baseToModelProto(m, cat.status)
					if err != nil {
						return status.Errorf(codes.Internal, "to proto: %s", err)
					}
					modelProtos = append(modelProtos, mp)
				}
			} else {
				var ms []*store.Model
				ms, more, err = store.ListModelsByActivationStatusWithPaginationInTransaction(
					tx,
					userInfo.ProjectID,
					true, /* onlyPublished */
					cat.status,
					afterID,
					int(limit)-len(modelProtos),
					req.IncludeLoadingModels,
				)
				if err != nil {
					return status.Errorf(codes.Internal, "list models: %s", err)
				}
				for _, m := range ms {
					modelProtos = append(modelProtos, toModelProto(m, cat.status))
				}
			}
			if more {
				hasMore = true
				break
			}

			if len(modelProtos) == int(limit) {
				endCat = i
				break
			}

			// Reset afterID for the next category.
			afterID = ""
		}

		if hasMore || endCat == -1 {
			return nil
		}

		// See if there are more models in future categories.
		for i := endCat; i < len(categories); i++ {
			if cat := categories[i]; cat.isBase {
				bms, _, err := store.ListBaseModelsByActivationStatusWithPaginationInTransaction(
					tx,
					userInfo.TenantID,
					cat.status,
					"", /* afterModelID */
					1,  /* limit */
					req.IncludeLoadingModels,
				)
				if err != nil {
					return status.Errorf(codes.Internal, "list base models: %s", err)
				}
				if len(bms) > 0 {
					hasMore = true
					break
				}
			} else {
				ms, _, err := store.ListModelsByActivationStatusWithPaginationInTransaction(
					tx,
					userInfo.ProjectID,
					true, /* onlyPublished */
					cat.status,
					"", /* afterModelID */
					1,  /* limit */
					req.IncludeLoadingModels,
				)
				if err != nil {
					return status.Errorf(codes.Internal, "list models: %s", err)
				}
				if len(ms) > 0 {
					hasMore = true
					break
				}
			}
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

func (s *S) findBaseModelOrModel(modelID, tenantID string, includeLoadingModels bool) (*store.BaseModel, *store.Model, error) {
	// Find a corresponding base model or a fine-tuned model
	var err error
	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: tenantID,
	}
	bm, err := s.store.GetBaseModel(k)
	if err == nil {
		if !isBaseModelLoaded(bm) && !includeLoadingModels {
			return nil, nil, status.Errorf(codes.InvalidArgument, "base model is not loaded")
		}
		return bm, nil, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	// Try a fine-tuned model next.
	m, err := s.store.GetModelByModelIDAndTenantID(modelID, tenantID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, status.Errorf(codes.Internal, "get model: %s", err)
		}

		return nil, nil, status.Errorf(codes.InvalidArgument, "invalid after: %s", err)
	}

	if !isModelLoaded(m) && !includeLoadingModels {
		return nil, nil, status.Errorf(codes.InvalidArgument, "model is not loaded")
	}

	return nil, m, nil
}

func (s *S) getTotalItems(userInfo *auth.UserInfo) (int32, error) {
	totalModels, err := s.store.CountModelsByProjectID(userInfo.ProjectID, true)
	if err != nil {
		return 0, err
	}

	totalBaseModels, err := s.store.CountBaseModels(userInfo.TenantID)
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

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: userInfo.TenantID,
	}
	m, err := getModel(s.store, k, req.IncludeLoadingModel)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func getModel(st *store.S, k store.ModelKey, includeLoadingModel bool) (*v1.Model, error) {
	// First check if it's a base model.
	bm, err := st.GetBaseModel(k)
	if err == nil {
		if !isBaseModelLoaded(bm) && !includeLoadingModel {
			return nil, status.Errorf(codes.NotFound, "model %q not found", k.ModelID)
		}

		as, err := getModelActivationStatus(st, k)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "get model activation status: %s", err)
		}

		mp, err := baseToModelProto(bm, as)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "to proto: %s", err)
		}
		return mp, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get base model: %s", err)
	}

	// Then check if it's a generated model.
	m, err := st.GetPublishedModelByModelIDAndTenantID(k.ModelID, k.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", k.ModelID)
		}
		return nil, status.Errorf(codes.Internal, "get published model by model id and tenant id: %s", err)
	}

	as, err := getModelActivationStatus(st, k)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get model activation status: %s", err)
	}

	return toModelProto(m, as), nil
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

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: userInfo.TenantID,
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

func (s *S) deleteFineTunedModel(
	ctx context.Context,
	k store.ModelKey,
) (*v1.DeleteModelResponse, error) {
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

func (s *S) deleteBaseModel(
	ctx context.Context,
	k store.ModelKey,
) (*v1.DeleteModelResponse, error) {
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
func (s *S) ActivateModel(
	ctx context.Context,
	req *v1.ActivateModelRequest,
) (*v1.ActivateModelResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: userInfo.TenantID,
	}
	if _, err := getModel(s.store, k, false /* includeLoadingModel */); err != nil {
		return nil, err
	}

	if err := s.store.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE); err != nil {
		return nil, status.Errorf(codes.Internal, "update model activation status: %s", err)
	}

	return &v1.ActivateModelResponse{}, nil
}

// DeactivateModel deactivates a model.
func (s *S) DeactivateModel(
	ctx context.Context,
	req *v1.DeactivateModelRequest,
) (*v1.DeactivateModelResponse, error) {
	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to extract user info from context")
	}

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: userInfo.TenantID,
	}
	if _, err := getModel(s.store, k, false /* includeLoadingModel */); err != nil {
		return nil, err
	}

	if err := s.store.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE); err != nil {
		return nil, status.Errorf(codes.Internal, "update model activation status: %s", err)
	}

	return &v1.DeactivateModelResponse{}, nil
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

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: clusterInfo.TenantID,
	}
	m, err := getModel(s.store, k, true /* includeLoadingModel */)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// ListModels lists models.
func (s *WS) ListModels(ctx context.Context, req *v1.ListModelsRequest) (*v1.ListModelsResponse, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bms, err := s.store.ListBaseModelsByTenantID(clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list base models: %s", err)
	}

	var modelProtos []*v1.Model
	for _, m := range bms {
		if !isBaseModelLoaded(m) {
			continue
		}

		k := store.ModelKey{
			ModelID:  m.ModelID,
			TenantID: m.TenantID,
		}
		as, err := getModelActivationStatus(s.store, k)
		if err != nil {
			return nil, err
		}
		mp, err := baseToModelProto(m, as)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "to proto: %s", err)
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

		k := store.ModelKey{
			ModelID:  m.ModelID,
			TenantID: m.TenantID,
		}
		as, err := getModelActivationStatus(s.store, k)
		if err != nil {
			return nil, err
		}
		modelProtos = append(modelProtos, toModelProto(m, as))
	}

	return &v1.ListModelsResponse{
		Object:     "list",
		Data:       modelProtos,
		HasMore:    false,
		TotalItems: int32(len(modelProtos)),
	}, nil
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
		path = fmt.Sprintf("%s/%s/%s", sc.PathPrefix, clusterInfo.TenantID, id)
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
		ModelID:  req.Id,
		TenantID: clusterInfo.TenantID,
	}
	existing, err := s.store.GetBaseModel(k)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Internal, "get base model: %s", err)
		}

		// Create a new base model.
		var m *store.BaseModel
		if err := s.store.Transaction(func(tx *gorm.DB) error {
			k := store.ModelKey{
				ModelID:  req.Id,
				TenantID: clusterInfo.TenantID,
			}
			var err error
			m, err = store.CreateBaseModelInTransaction(tx, k, req.Path, formats, req.GgufModelPath, req.SourceRepository)
			if err != nil {
				return status.Errorf(codes.Internal, "create base model: %s", err)
			}

			if err := store.CreateModelActivationStatusInTransaction(tx, &store.ModelActivationStatus{
				ModelID:  req.Id,
				TenantID: clusterInfo.TenantID,
				Status:   v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
			}); err != nil {
				return status.Errorf(codes.Internal, "create model activation status: %s", err)
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

	k := store.ModelKey{
		ModelID:  req.Id,
		TenantID: clusterInfo.TenantID,
	}
	m, err := s.store.GetBaseModel(k)
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

	if req.LoadingResult == nil {
		return nil, status.Error(codes.InvalidArgument, "loading_result is required")
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

	if req.LoadingResult == nil {
		return nil, status.Error(codes.InvalidArgument, "loading_result is required")
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
		return nil, status.Error(codes.InvalidArgument, "invalid loading_result")
	}

	if err != nil {
		if errors.Is(err, store.ErrConcurrentUpdate) {
			return nil, status.Errorf(codes.FailedPrecondition, "concurrent update to model status")
		}

		return nil, status.Errorf(codes.Internal, "update model loading status: %s", err)
	}

	return &v1.UpdateModelLoadingStatusResponse{}, nil
}

func (s *WS) generateModelID(baseModel, suffix, tenantID string) (string, error) {
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
		if _, err := s.store.GetModelByModelIDAndTenantID(id, tenantID); errors.Is(err, gorm.ErrRecordNotFound) {
			return id, nil
		}
	}
}

func toModelProto(m *store.Model, as v1.ActivationStatus) *v1.Model {
	return &v1.Model{
		Id:                   m.ModelID,
		Object:               "model",
		Created:              m.CreatedAt.UTC().Unix(),
		OwnedBy:              "user",
		LoadingStatus:        toLoadingStatus(m.LoadingStatus),
		SourceRepository:     v1.SourceRepository_SOURCE_REPOSITORY_FINE_TUNING,
		LoadingFailureReason: m.LoadingFailureReason,
		// Fine-tuned models always have the Hugging Face format.
		Formats:     []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		IsBaseModel: false,
		BaseModelId: m.BaseModelID,

		ActivationStatus: as,
	}
}

func baseToModelProto(m *store.BaseModel, as v1.ActivationStatus) (*v1.Model, error) {
	formats, err := store.UnmarshalModelFormats(m.Formats)
	if err != nil {
		return nil, err
	}
	if len(formats) == 0 {
		// For backward compatibility.
		formats = []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}
	}

	return &v1.Model{
		Id:                   m.ModelID,
		Object:               "model",
		Created:              m.CreatedAt.UTC().Unix(),
		OwnedBy:              "system",
		LoadingStatus:        toLoadingStatus(m.LoadingStatus),
		SourceRepository:     m.SourceRepository,
		LoadingFailureReason: m.LoadingFailureReason,
		Formats:              formats,
		IsBaseModel:          true,
		BaseModelId:          "",

		ActivationStatus: as,
	}, nil
}

func toBaseModelProto(m *store.BaseModel) *v1.BaseModel {
	return &v1.BaseModel{
		Id:      m.ModelID,
		Object:  "basemodel",
		Created: m.CreatedAt.UTC().Unix(),
	}
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
