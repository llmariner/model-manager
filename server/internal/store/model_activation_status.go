package store

import (
	"gorm.io/gorm"

	v1 "github.com/llmariner/model-manager/api/v1"
)

// ModelActivationStatus represents a model activation status.
type ModelActivationStatus struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_model_activation_status_model_id_tenant_id"`
	ModelID  string `gorm:"uniqueIndex:idx_model_activation_status_model_id_tenant_id"`

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
func (s *S) GetModelActivationStatus(modelID, tenantID string) (*ModelActivationStatus, error) {
	return GetModelActivationStatusInTransaction(s.db, modelID, tenantID)
}

// GetModelActivationStatusInTransaction retrieves the model activation status in a transaction.
func GetModelActivationStatusInTransaction(tx *gorm.DB, modelID, tenantID string) (*ModelActivationStatus, error) {
	status := &ModelActivationStatus{}
	if err := tx.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).First(status).Error; err != nil {
		return nil, err
	}

	return status, nil
}

// UpdateModelActivationStatus updates the model activation status.
func (s *S) UpdateModelActivationStatus(modelID, tenantID string, status v1.ActivationStatus) error {
	res := s.db.Model(&ModelActivationStatus{}).
		Where("model_id = ? AND tenant_id = ?", modelID, tenantID).
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
func (s *S) DeleteModelActivationStatus(modelID, tenantID string) error {
	return DeleteModelActivationStatusInTransaction(s.db, modelID, tenantID)
}

// DeleteModelActivationStatusInTransaction deletes the model activation status.
func DeleteModelActivationStatusInTransaction(tx *gorm.DB, modelID, tenantID string) error {
	res := tx.Unscoped().Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&ModelActivationStatus{})
	if err := res.Error; err != nil {
		return err
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
