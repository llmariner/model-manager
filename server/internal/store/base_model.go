package store

import (
	"gorm.io/gorm"
)

// BaseModel represents a base model.
type BaseModel struct {
	gorm.Model

	ModelID string `gorm:"uniqueIndex"`
	Path    string

	// GGUFModelPath is the path to the GGUF model.
	GGUFModelPath string
}

// CreateBaseModel creates a model.
func (s *S) CreateBaseModel(modelID, path, ggufModelPath string) (*BaseModel, error) {
	m := &BaseModel{
		ModelID:       modelID,
		Path:          path,
		GGUFModelPath: ggufModelPath,
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

// ListBaseModels returns all base models.
func (s *S) ListBaseModels() ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}
