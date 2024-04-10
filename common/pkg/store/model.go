package store

import (
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model

	ModelID  string `gorm:"uniqueIndex:idx_model_model_id_tenant_id"`
	TenantID string `gorm:"uniqueIndex:idx_model_model_id_tenant_id"`

	// TODO(kenji): Add a model location.
}

// ModelKey represents a model key.
type ModelKey struct {
	ModelID  string
	TenantID string
}

// CreateModel creates a model.
func (s *S) CreateModel(k ModelKey) (*Model, error) {
	m := &Model{
		ModelID:  k.ModelID,
		TenantID: k.TenantID,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GeteModel returns a model by model ID and tenant ID.
func (s *S) GetModel(k ModelKey) (*Model, error) {
	var m Model
	if err := s.db.Where("model_id = ? AND tenant_id = ?", k.ModelID, k.TenantID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListModelsByTenantID finds models.
func (s *S) ListModelsByTenantID(tenantID string) ([]*Model, error) {
	var ms []*Model
	if err := s.db.Where("tenant_id = ?", tenantID).Find(&ms).Error; err != nil {
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
