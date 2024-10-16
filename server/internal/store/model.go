package store

import (
	"github.com/llmariner/model-manager/api/v1"
	"gorm.io/gorm"
)

// Model represents a model.
type Model struct {
	gorm.Model

	// ModelID is the model ID. It is globally unique.
	ModelID string `gorm:"uniqueIndex"`

	TenantID       string `gorm:"index"`
	OrganizationID string
	ProjectID      string `gorm:"index"`

	Path        string
	IsPublished bool

	BaseModelID  string `gorm:"index"`
	Adapter      v1.AdapterType
	Quantization v1.QuantizationType
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
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GetPublishedModelByModelIDAndProjectID returns a published model by model ID and tenant ID.
func (s *S) GetPublishedModelByModelIDAndProjectID(modelID, projectID string) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND project_id = ? AND is_published = ? ", modelID, projectID, true).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetModelByModelID returns a model by model ID.
func (s *S) GetModelByModelID(modelID string) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ?", modelID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetPublishedModelByModelIDAndTenantID returns a model by model ID.
func (s *S) GetPublishedModelByModelIDAndTenantID(modelID, tenantID string) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND tenant_id = ? AND is_published = ?", modelID, tenantID, true).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListModelsByProjectID finds models.
func (s *S) ListModelsByProjectID(projectID string, onlyPublished bool) ([]*Model, error) {
	q := s.db.Where("project_id = ?", projectID)
	if onlyPublished {
		q = q.Where("is_published = true")
	}
	q = q.Order("id DESC")

	var ms []*Model
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// DeleteModel deletes a model by model ID and tenant ID.
func (s *S) DeleteModel(modelID, projectID string) error {
	res := s.db.Unscoped().Where("model_id = ? AND project_id = ?", modelID, projectID).Delete(&Model{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateModel updates the model.
func (s *S) UpdateModel(modelID string, tenantID string, isPublished bool) error {
	res := s.db.Model(&Model{}).Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Update("is_published", isPublished)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
