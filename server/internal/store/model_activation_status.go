package store

import (
	"gorm.io/gorm"

	v1 "github.com/llmariner/model-manager/api/v1"
)

// ModelActivationStatus represents a model activation status.
type ModelActivationStatus struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_model_activation_status_model_id_tenant_id_project_id"`
	ModelID  string `gorm:"uniqueIndex:idx_model_activation_status_model_id_tenant_id_project_id"`

	// ProjectID is set only for project-scoped base models. This is set to empty
	// for global-scoped base models and fine-tuned models. (Fine-tuned models have emtpy values
	// for backward compatibility.)
	ProjectID string `gorm:"uniqueIndex:idx_model_activation_status_model_id_tenant_id_project_id"`

	Status v1.ActivationStatus
}

// CreateModelActivationStatus creates a new model activation status.
func (s *S) CreateModelActivationStatus(status *ModelActivationStatus) error {
	return CreateModelActivationStatusInTransaction(s.db, status)
}

// CreateModelActivationStatusInTransaction creates a new model activation status in a transaction.
func CreateModelActivationStatusInTransaction(tx *gorm.DB, status *ModelActivationStatus) error {
	if err := tx.Create(status).Error; err != nil {
		return err
	}

	return nil
}

// GetModelActivationStatus retrieves the model activation status.
func (s *S) GetModelActivationStatus(k ModelKey) (*ModelActivationStatus, error) {
	return GetModelActivationStatusInTransaction(s.db, k)
}

// GetModelActivationStatusInTransaction retrieves the model activation status in a transaction.
func GetModelActivationStatusInTransaction(tx *gorm.DB, k ModelKey) (*ModelActivationStatus, error) {
	status := &ModelActivationStatus{}
	if err := k.buildQuery(tx).First(status).Error; err != nil {
		return nil, err
	}

	return status, nil
}

// UpdateModelActivationStatus updates the model activation status.
func (s *S) UpdateModelActivationStatus(k ModelKey, status v1.ActivationStatus) error {
	res := k.buildQuery(s.db.Model(&ModelActivationStatus{})).
		Update("status", status)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteModelActivationStatus deletes the model activation status.
func (s *S) DeleteModelActivationStatus(k ModelKey) error {
	return DeleteModelActivationStatusInTransaction(s.db, k)
}

// DeleteModelActivationStatusInTransaction deletes the model activation status.
func DeleteModelActivationStatusInTransaction(tx *gorm.DB, k ModelKey) error {
	res := k.buildQuery(tx.Unscoped()).Delete(&ModelActivationStatus{})
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
