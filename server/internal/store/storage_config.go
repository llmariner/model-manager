package store

import (
	"gorm.io/gorm"
)

// StorageConfig represents a storage configuration for a tenant.
type StorageConfig struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex"`

	// PathPrefix is the prefix of the S3 path for storing models.
	PathPrefix string
}

// CreateStorageConfig creates a storage configuration.
func (s *S) CreateStorageConfig(tenantID, pathPrefix string) (*StorageConfig, error) {
	c := &StorageConfig{
		TenantID:   tenantID,
		PathPrefix: pathPrefix,
	}
	if err := s.db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// GetStorageConfig returns a storage configuration by tenant ID.
func (s *S) GetStorageConfig(tenantID string) (*StorageConfig, error) {
	var c StorageConfig
	if err := s.db.Where("tenant_id = ?", tenantID).Take(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
