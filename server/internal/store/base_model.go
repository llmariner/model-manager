package store

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	v1 "github.com/llmariner/model-manager/api/v1"
)

// BaseModel represents a base model.
type BaseModel struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id_project_id"`

	ModelID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id_project_id"`

	// ProjectID is the ID of the project to which the model belongs. It is empty if
	// the model is globally scoped and it is not associated with any project.
	ProjectID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id_project_id"`

	Path string

	Formats []byte

	// GGUFModelPath is the path to the GGUF model.
	GGUFModelPath string

	SourceRepository v1.SourceRepository
	LoadingStatus    v1.ModelLoadingStatus

	LoadingFailureReason string
}

// UnmarshalModelFormats unmarshals model formats.
func UnmarshalModelFormats(b []byte) ([]v1.ModelFormat, error) {
	var formats v1.ModelFormats
	if err := proto.Unmarshal(b, &formats); err != nil {
		return nil, err
	}
	return formats.Formats, nil
}

// CreateBaseModel creates a model.
func (s *S) CreateBaseModel(
	k ModelKey,
	path string,
	formats []v1.ModelFormat,
	ggufModelPath string,
	sourceRepository v1.SourceRepository,
) (*BaseModel, error) {
	return CreateBaseModelInTransaction(
		s.db,
		k,
		path,
		formats,
		ggufModelPath,
		sourceRepository,
	)
}

// CreateBaseModelInTransaction creates a model in a transaction.
func CreateBaseModelInTransaction(
	tx *gorm.DB,
	k ModelKey,
	path string,
	formats []v1.ModelFormat,
	ggufModelPath string,
	sourceRepository v1.SourceRepository,
) (*BaseModel, error) {
	b, err := marshalFormats(formats)
	if err != nil {
		return nil, err
	}

	m := &BaseModel{
		ModelID:          k.ModelID,
		ProjectID:        k.ProjectID,
		Path:             path,
		Formats:          b,
		GGUFModelPath:    ggufModelPath,
		LoadingStatus:    v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		SourceRepository: sourceRepository,
		TenantID:         k.TenantID,
	}
	if err := tx.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// CreateBaseModelWithLoadingRequested creates a model with the requested loading status.
func (s *S) CreateBaseModelWithLoadingRequested(
	k ModelKey,
	sourceRepository v1.SourceRepository,
) (*BaseModel, error) {
	return CreateBaseModelWithLoadingRequestedInTransaction(
		s.db,
		k,
		sourceRepository,
	)
}

// CreateBaseModelWithLoadingRequestedInTransaction creates a model with the requested loading status in a transaction.
func CreateBaseModelWithLoadingRequestedInTransaction(
	tx *gorm.DB,
	k ModelKey,
	sourceRepository v1.SourceRepository,
) (*BaseModel, error) {
	m := &BaseModel{
		ModelID:          k.ModelID,
		ProjectID:        k.ProjectID,
		SourceRepository: sourceRepository,
		LoadingStatus:    v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
		TenantID:         k.TenantID,
	}
	if err := tx.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// DeleteBaseModel deletes a base model by model ID and tenant ID.
func (s *S) DeleteBaseModel(k ModelKey) error {
	return DeleteBaseModelInTransaction(s.db, k)
}

// DeleteBaseModelInTransaction deletes a base model.
func DeleteBaseModelInTransaction(tx *gorm.DB, k ModelKey) error {
	res := k.buildQuery(tx.Unscoped()).Delete(&BaseModel{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetBaseModel returns a base model by a model key.
func (s *S) GetBaseModel(k ModelKey) (*BaseModel, error) {
	var m BaseModel
	if err := k.buildQuery(s.db).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListBaseModelsByTenantID returns all base models for a tenant.
func (s *S) ListBaseModelsByTenantID(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? ", tenantID).Order("model_id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListBaseModelsByModelIDAndTenantID returns all base models that has a given ID for a tenant,
// including both global-scoped ones and project-scoped ones.
func (s *S) ListBaseModelsByModelIDAndTenantID(modelID, tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.
		Where("model_id = ? AND tenant_id = ? ", modelID, tenantID).
		Order("project_id DESC").
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListBaseModelsByActivationStatusWithPaginationInTransaction finds base models filtered by activation status with pagination in a transaction.
// Models are returned with an ascending order of model IDs.
func ListBaseModelsByActivationStatusWithPaginationInTransaction(
	tx *gorm.DB,
	projectID string,
	tenantID string,
	status v1.ActivationStatus,
	afterModelID string,
	limit int,
	includeLoadingModels bool,
) ([]*BaseModel, bool, error) {
	// We need to handle a case where the same base model exists both with a global scope and a project scope.
	// In that case, we want to return the project-scoped one with the following steps.
	//
	// 1. Get distinct model IDs that match the conditions.
	// 2. Query project-scoped models that have those model IDs.
	// 3. Query global-scoped models that have those model IDs if no project-scoped models are found.

	var modelIDs []string
	q := tx.Table("base_models AS bm").
		Select("bm.model_id").
		Distinct("bm.model_id").
		Joins("JOIN model_activation_statuses AS mas ON mas.model_id = bm.model_id AND mas.tenant_id = bm.tenant_id").
		// Find all models that are either globally scoped (project_id is NULL) or
		// scoped to the given project.
		Where("(bm.project_id IS NULL OR bm.project_id = '' OR bm.project_id = ?)", projectID).
		Where("bm.tenant_id = ?", tenantID).
		Where("mas.status = ?", status)
	if afterModelID != "" {
		q = q.Where("bm.model_id > ?", afterModelID)
	}
	if !includeLoadingModels {
		q = q.Where("(bm.loading_status IS NULL OR bm.loading_status = ?)", v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)
	}
	if err := q.
		Order("bm.model_id").
		Limit(limit + 1).
		Find(&modelIDs).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(modelIDs) > limit {
		modelIDs = modelIDs[:limit]
		hasMore = true
	}

	// Query project-scoped models that have the model IDs.
	var pmodels []*BaseModel
	if err := tx.
		Where("project_id = ?", projectID).
		Where("model_id IN ?", modelIDs).
		Where("tenant_id = ?", tenantID).
		Find(&pmodels).Error; err != nil {
		return nil, false, err
	}

	foundIDs := make(map[string]bool)
	for _, m := range pmodels {
		foundIDs[m.ModelID] = true
	}
	var remainingIDs []string
	for _, id := range modelIDs {
		if foundIDs[id] {
			continue
		}
		remainingIDs = append(remainingIDs, id)
	}

	// Query global-scoped models that have the model IDs.
	var gmodels []*BaseModel
	if err := tx.
		Where("(project_id IS NULL OR project_id = '')").
		Where("model_id IN ?", remainingIDs).
		Where("tenant_id = ?", tenantID).
		Find(&gmodels).Error; err != nil {
		return nil, false, err
	}

	return append(pmodels, gmodels...), hasMore, nil
}

// ListUnloadedBaseModels returns all unloaded base models with the requested loading status.
func (s *S) ListUnloadedBaseModels(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? AND loading_status = ?", tenantID, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED).
		Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListLoadingBaseModels returns all base models with the loading status.
func (s *S) ListLoadingBaseModels(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? AND loading_status = ?", tenantID, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING).
		Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// updateBaseModel updates the model if the current status matches with the given one.
func (s *S) updateBaseModel(
	k ModelKey,
	curr v1.ModelLoadingStatus,
	updates map[string]interface{},
) error {
	res := k.buildQuery(s.db.Model(&BaseModel{})).
		Where("loading_status = ?", curr).
		Updates(updates)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

// UpdateBaseModelToLoadingStatus updates the loading status to LOADING.
func (s *S) UpdateBaseModelToLoadingStatus(k ModelKey) error {
	return s.updateBaseModel(
		k,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
		map[string]interface{}{
			"loading_status": v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		},
	)
}

// UpdateBaseModelToSucceededStatus updates the loading status to SUCCEEDED and updates other relevant information.
func (s *S) UpdateBaseModelToSucceededStatus(
	k ModelKey,
	path string,
	formats []v1.ModelFormat,
	ggufModelPath string,
) error {
	b, err := marshalFormats(formats)
	if err != nil {
		return err
	}

	return s.updateBaseModel(
		k,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"path":            path,
			"formats":         b,
			"gguf_model_path": ggufModelPath,
			"loading_status":  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		},
	)
}

// UpdateBaseModelToFailedStatus updates the loading status to FAILED and updates other relevant information.
func (s *S) UpdateBaseModelToFailedStatus(k ModelKey, failureReason string) error {
	return s.updateBaseModel(
		k,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"loading_failure_reason": failureReason,
			"loading_status":         v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED,
		},
	)
}

// CountBaseModels counts the total number of base models.
func (s *S) CountBaseModels(projectID, tenantID string) (int64, error) {
	var count int64
	if err := s.db.Model(&BaseModel{}).
		Distinct("model_id").
		// Find all models that are either globally scoped (project_id is NULL) or
		// scoped to the given project.
		Where("(project_id IS NULL OR project_id = '' OR project_id = ?)", projectID).
		Where("tenant_id = ? ", tenantID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func marshalFormats(formats []v1.ModelFormat) ([]byte, error) {
	p := v1.ModelFormats{
		Formats: formats,
	}
	return proto.Marshal(&p)
}
