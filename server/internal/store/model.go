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

	LoadingStatus v1.ModelLoadingStatus
}

// ModelSpec represents a model spec that is passed to CreateModel.
type ModelSpec struct {
	ModelID        string
	TenantID       string
	OrganizationID string
	ProjectID      string
	Path           string
	IsPublished    bool
	BaseModelID    string
	Adapter        v1.AdapterType
	Quantization   v1.QuantizationType
	LoadingStatus  v1.ModelLoadingStatus
}

// CreateModel creates a model.
func (s *S) CreateModel(spec ModelSpec) (*Model, error) {
	m := &Model{
		ModelID:        spec.ModelID,
		TenantID:       spec.TenantID,
		OrganizationID: spec.OrganizationID,
		ProjectID:      spec.ProjectID,
		Path:           spec.Path,
		IsPublished:    spec.IsPublished,
		BaseModelID:    spec.BaseModelID,
		Adapter:        spec.Adapter,
		Quantization:   spec.Quantization,
		LoadingStatus:  spec.LoadingStatus,
	}
	if err := s.db.Create(m).Error; err != nil {
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

// DeleteModel deletes a model by model ID and tenant ID.
func (s *S) DeleteModel(modelID, tenantID string) error {
	res := s.db.Unscoped().Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&Model{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateModel updates the model.
func (s *S) UpdateModel(modelID string, tenantID string, isPublished bool, loadingStatus v1.ModelLoadingStatus) error {
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
