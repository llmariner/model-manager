package store

import (
	"gorm.io/gorm"
)

// BaseModel represents a base model.
type BaseModel struct {
	gorm.Model

	ModelID string `gorm:"uniqueIndex"`
	Path    string
}

// CreateBaseModel creates a model.
func (s *S) CreateBaseModel(modelID, path string) (*BaseModel, error) {
	m := &BaseModel{
		ModelID: modelID,
		Path:    path,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// GetBaseModel returns a base model by model ID.
func (s *S) GetBaseModel(modelID string) (*BaseModel, error) {
	var m BaseModel
	if err := s.db.Where("model_id = ?", modelID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
