package store

import (
	"gorm.io/gorm"
)

// Model represents a model.
type Model struct {
	gorm.Model

	// ModelID is the model ID. It is globally unique.
	ModelID string `gorm:"uniqueIndex"`

	TenantID string `gorm:"index"`

	Path        string
	IsPublished bool
}

// ModelKey represents a model key.
type ModelKey struct {
	ModelID  string
	TenantID string
}

// ModelSpec represents a model spec that is passed to CreateModel.
type ModelSpec struct {
	Key         ModelKey
	Path        string
	IsPublished bool
}

// CreateModel creates a model.
func (s *S) CreateModel(spec ModelSpec) (*Model, error) {
	m := &Model{
		ModelID:     spec.Key.ModelID,
		TenantID:    spec.Key.TenantID,
		Path:        spec.Path,
		IsPublished: spec.IsPublished,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GetPublishedModel returns a published model by model ID and tenant ID.
func (s *S) GetPublishedModel(k ModelKey) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND tenant_id = ? AND is_published = ? ", k.ModelID, k.TenantID, true).Take(&m).Error; err != nil {
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

// ListModelsByTenantID finds models.
func (s *S) ListModelsByTenantID(tenantID string, onlyPublished bool) ([]*Model, error) {
	q := s.db.Where("tenant_id = ?", tenantID)
	if onlyPublished {
		q = q.Where("is_published = true")
	}

	var ms []*Model
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListAllPublishedModels finds all published models.
func (s *S) ListAllPublishedModels() ([]*Model, error) {
	var ms []*Model
	if err := s.db.Where("is_published = true").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// DeleteModel deletes a model by model ID and tenant ID.
func (s *S) DeleteModel(k ModelKey) error {
	res := s.db.Unscoped().Where("model_id = ? AND tenant_id = ?", k.ModelID, k.TenantID).Delete(&Model{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateModel updates the model.
func (s *S) UpdateModel(k ModelKey, isPublished bool) error {
	res := s.db.Model(&Model{}).Where("model_id = ? AND tenant_id = ?", k.ModelID, k.TenantID).Update("is_published", isPublished)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
