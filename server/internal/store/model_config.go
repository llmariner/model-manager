package store

import (
	"gorm.io/gorm"
)

// ModelConfig represents a model configuration.
type ModelConfig struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_model_config_model_id_tenant_id_project_id"`
	ModelID  string `gorm:"uniqueIndex:idx_model_config_model_id_tenant_id_project_id"`

	// ProjectID is set only for project-scoped base models. This is set to empty
	// for global-scoped base models and fine-tuned models.
	ProjectID string `gorm:"uniqueIndex:idx_model_config_model_id_tenant_id_project_id"`

	// EncodedConfig is a proto-encoded message that contains v1.ModelConfig.
	EncodedConfig []byte
}

// CreateModelConfig creates a new model config.
func (s *S) CreateModelConfig(c *ModelConfig) error {
	return CreateModelConfigInTransaction(s.db, c)
}

// CreateModelConfigInTransaction creates a new model config in a transaction.
func CreateModelConfigInTransaction(tx *gorm.DB, c *ModelConfig) error {
	if err := tx.Create(c).Error; err != nil {
		return err
	}

	return nil
}

// GetModelConfig retrieves the model config.
func (s *S) GetModelConfig(k ModelKey) (*ModelConfig, error) {
	return GetModelConfigInTransaction(s.db, k)
}

// GetModelConfigInTransaction retrieves the model config in a transaction.
func GetModelConfigInTransaction(tx *gorm.DB, k ModelKey) (*ModelConfig, error) {
	c := &ModelConfig{}
	if err := k.buildQuery(tx).First(c).Error; err != nil {
		return nil, err
	}

	return c, nil
}

// UpdateModelConfig updates the model config.
func (s *S) UpdateModelConfig(k ModelKey, encodedConfig []byte) error {
	return UpdateModelConfigInTransaction(s.db, k, encodedConfig)
}

// UpdateModelConfigInTransaction updates the model config.
func UpdateModelConfigInTransaction(tx *gorm.DB, k ModelKey, encodedConfig []byte) error {
	res := k.buildQuery(tx.Model(&ModelConfig{})).
		Update("encoded_config", encodedConfig)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteModelConfig deletes the model config.
func (s *S) DeleteModelConfig(k ModelKey) error {
	return DeleteModelConfigInTransaction(s.db, k)
}

// DeleteModelConfigInTransaction deletes the model config.
func DeleteModelConfigInTransaction(tx *gorm.DB, k ModelKey) error {
	res := k.buildQuery(tx.Unscoped()).Delete(&ModelConfig{})
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
