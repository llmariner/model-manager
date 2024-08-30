package store

import (
	"gorm.io/gorm"

	v1 "github.com/llm-operator/model-manager/api/v1"
)

// BaseModel represents a base model.
type BaseModel struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id"`

	ModelID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id"`
	Path    string

	Format v1.ModelFormat

	// GGUFModelPath is the path to the GGUF model.
	GGUFModelPath string
}

// CreateBaseModel creates a model.
func (s *S) CreateBaseModel(
	modelID string,
	path string,
	format v1.ModelFormat,
	ggufModelPath string,
	tenantID string,
) (*BaseModel, error) {
	m := &BaseModel{
		ModelID:       modelID,
		Path:          path,
		Format:        format,
		GGUFModelPath: ggufModelPath,
		TenantID:      tenantID,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GetBaseModel returns a base model by model ID and tenant ID.
func (s *S) GetBaseModel(modelID, tenantID string) (*BaseModel, error) {
	var m BaseModel
	if err := s.db.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListBaseModels returns all base models for a tenant.
func (s *S) ListBaseModels(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? ", tenantID).Order("id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}
