package store

import (
	v1 "github.com/llmariner/model-manager/api/v1"
	"gorm.io/gorm"
)

// Model represents a model.
type Model struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_model_model_id_tenant_id"`
	ModelID  string `gorm:"uniqueIndex:idx_model_model_id_tenant_id"`

	OrganizationID string
	ProjectID      string `gorm:"index"`

	Path        string
	IsPublished bool

	BaseModelID  string `gorm:"index"`
	Adapter      v1.AdapterType
	Quantization v1.QuantizationType

	LoadingStatus        v1.ModelLoadingStatus
	LoadingFailureReason string

	SourceRepository  v1.SourceRepository
	ModelFileLocation string
}

// ModelSpec represents a model spec that is passed to CreateModel.
type ModelSpec struct {
	ModelID           string
	TenantID          string
	OrganizationID    string
	ProjectID         string
	Path              string
	IsPublished       bool
	BaseModelID       string
	Adapter           v1.AdapterType
	Quantization      v1.QuantizationType
	LoadingStatus     v1.ModelLoadingStatus
	SourceRepository  v1.SourceRepository
	ModelFileLocation string
}

// CreateModel creates a model.
func (s *S) CreateModel(spec ModelSpec) (*Model, error) {
	return CreateModelInTransaction(s.db, spec)
}

// CreateModelInTransaction creates a model in a transaction.
func CreateModelInTransaction(tx *gorm.DB, spec ModelSpec) (*Model, error) {
	m := &Model{
		ModelID:           spec.ModelID,
		TenantID:          spec.TenantID,
		OrganizationID:    spec.OrganizationID,
		ProjectID:         spec.ProjectID,
		Path:              spec.Path,
		IsPublished:       spec.IsPublished,
		BaseModelID:       spec.BaseModelID,
		Adapter:           spec.Adapter,
		Quantization:      spec.Quantization,
		LoadingStatus:     spec.LoadingStatus,
		SourceRepository:  spec.SourceRepository,
		ModelFileLocation: spec.ModelFileLocation,
	}
	if err := tx.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GetPublishedModelByModelIDAndTenantID returns a published model by model ID and tenant ID.
func (s *S) GetPublishedModelByModelIDAndTenantID(modelID, tenantID string) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND tenant_id = ? AND is_published = ? ", modelID, tenantID, true).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetModelByModelIDAndTenantID returns a model by model ID and tenant ID.
func (s *S) GetModelByModelIDAndTenantID(modelID, tenantID string) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListModelsByTenantID returns all models by tenant ID.
func (s *S) ListModelsByTenantID(tenantID string,
) ([]*Model, error) {
	var ms []*Model
	if err := s.db.Where("tenant_id = ?", tenantID).Order("model_id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListModelsByProjectIDWithPagination finds models with pagination. Models are returned with an ascending order of ID.
func (s *S) ListModelsByProjectIDWithPagination(
	projectID string,
	onlyPublished bool,
	afterModelID string,
	limit int,
	includeLoadingModels bool,
) ([]*Model, bool, error) {
	var ms []*Model
	q := s.db.Where("project_id = ?", projectID)
	if onlyPublished {
		q = q.Where("is_published = true")
	}
	if afterModelID != "" {
		q = q.Where("model_id > ?", afterModelID)
	}
	if !includeLoadingModels {
		q = q.Where("(loading_status is null OR loading_status = ?)", v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)
	}
	if err := q.Order("model_id").Limit(limit + 1).Find(&ms).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(ms) > limit {
		ms = ms[:limit]
		hasMore = true
	}
	return ms, hasMore, nil
}

// ListModelsByActivationStatusWithPagination finds models filtered by activation status with pagination.
// Models are returned with an ascending order of model IDs.
func (s *S) ListModelsByActivationStatusWithPagination(
	projectID string,
	onlyPublished bool,
	status v1.ActivationStatus,
	afterModelID string,
	limit int,
	includeLoadingModels bool,
) ([]*Model, bool, error) {
	return ListModelsByActivationStatusWithPaginationInTransaction(s.db, projectID, onlyPublished, status, afterModelID, limit, includeLoadingModels)
}

// ListModelsByActivationStatusWithPaginationInTransaction finds models filtered by activation status with pagination in a transaction.
// Models are returned with an ascending order of model IDs.
func ListModelsByActivationStatusWithPaginationInTransaction(
	tx *gorm.DB,
	projectID string,
	onlyPublished bool,
	status v1.ActivationStatus,
	afterModelID string,
	limit int,
	includeLoadingModels bool,
) ([]*Model, bool, error) {
	var ms []*Model
	q := tx.Joins("JOIN model_activation_statuses AS mas ON mas.model_id = models.model_id AND mas.tenant_id = models.tenant_id").
		Where("models.project_id = ? AND mas.status = ?", projectID, status)
	if onlyPublished {
		q = q.Where("models.is_published = true")
	}
	if afterModelID != "" {
		q = q.Where("models.model_id > ?", afterModelID)
	}
	if !includeLoadingModels {
		q = q.Where("(models.loading_status is null OR models.loading_status = ?)", v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)
	}
	if err := q.Order("models.model_id").Limit(limit + 1).Find(&ms).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(ms) > limit {
		ms = ms[:limit]
		hasMore = true
	}
	return ms, hasMore, nil
}

// ListUnloadedModels returns all unloaded base models with the requested loading status.
func (s *S) ListUnloadedModels(tenantID string) ([]*Model, error) {
	var ms []*Model
	if err := s.db.Where("tenant_id = ? AND loading_status = ?", tenantID, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED).
		Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// updateModel updates the model if the current status matches with the given one.
func (s *S) updateModel(
	modelID string,
	tenantID string,
	curr v1.ModelLoadingStatus,
	updates map[string]interface{},
) error {
	res := s.db.Model(&Model{}).
		Where("model_id = ? AND tenant_id = ? AND loading_status = ?", modelID, tenantID, curr).
		Updates(updates)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

// UpdateModelToLoadingStatus updates the loading status to LOADING.
func (s *S) UpdateModelToLoadingStatus(modelID string, tenantID string) error {
	return s.updateModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
		map[string]interface{}{
			"loading_status": v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		},
	)
}

// UpdateModelToSucceededStatus updates the loading status to SUCCEEDED and updates other relevant information.
func (s *S) UpdateModelToSucceededStatus(modelID string, tenantID string) error {
	return s.updateModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"loading_status": v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		},
	)
}

// UpdateModelToFailedStatus updates the loading status to FAILED and updates other relevant information.
func (s *S) UpdateModelToFailedStatus(modelID string, tenantID string, failureReason string) error {
	return s.updateModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"loading_failure_reason": failureReason,
			"loading_status":         v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED,
		},
	)
}

// DeleteModel deletes a model by model ID and tenant ID.
func (s *S) DeleteModel(modelID, tenantID string) error {
	return DeleteModelInTransaction(s.db, modelID, tenantID)
}

// DeleteModelInTransaction deletes a model by model ID and tenant ID in a transaction.
func DeleteModelInTransaction(tx *gorm.DB, modelID, tenantID string) error {
	res := tx.Unscoped().Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&Model{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateModelPublishingStatus updates the publishing status of a model.
func (s *S) UpdateModelPublishingStatus(modelID string, tenantID string, isPublished bool, loadingStatus v1.ModelLoadingStatus) error {
	res := s.db.Model(&Model{}).Where("model_id = ? AND tenant_id = ?", modelID, tenantID).
		Updates(map[string]interface{}{
			"is_published":   isPublished,
			"loading_status": loadingStatus,
		})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountModelsByProjectID counts the total number of models by project ID.
func (s *S) CountModelsByProjectID(projectID string, onlyPublished bool) (int64, error) {
	var count int64
	if err := s.db.Model(&Model{}).Where("project_id = ? AND is_published = ?", projectID, onlyPublished).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
