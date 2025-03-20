package store

import (
	"gorm.io/gorm"
)

// HFModelRepo represents a HuggingFace model repository where models are downloaded from.
// This is used to track a HuggingFace repo that has one-to-many mapping to base models.
//
// The record is created when the download completes.
type HFModelRepo struct {
	gorm.Model

	Name     string `gorm:"uniqueIndex:idx_hf_model_repo_name_tenant_id"`
	TenantID string `gorm:"uniqueIndex:idx_hf_model_repo_name_tenant_id"`
}

// CreateHFModelRepo creates a model repo.
func (s *S) CreateHFModelRepo(
	name string,
	tenantID string,
) (*HFModelRepo, error) {
	r := &HFModelRepo{
		Name:     name,
		TenantID: tenantID,
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// GetHFModelRepo returns a repo by the repo namen and tenant ID.
func (s *S) GetHFModelRepo(name, tenantID string) (*HFModelRepo, error) {
	var r HFModelRepo
	if err := s.db.Where("name = ? AND tenant_id = ?", name, tenantID).Take(&r).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

// ListHFModelRepos returns all HuggingFace model repos for a tenant.
func (s *S) ListHFModelRepos(tenantID string) ([]*HFModelRepo, error) {
	var rs []*HFModelRepo
	if err := s.db.Where("tenant_id = ? ", tenantID).Order("id DESC").Find(&rs).Error; err != nil {
		return nil, err
	}
	return rs, nil
}

// DeleteHFModelRepoInTransaction deletes a model repo.
func DeleteHFModelRepoInTransaction(tx *gorm.DB, name, tenantID string) error {
	res := tx.Where("name = ? AND tenant_id = ?", name, tenantID).Delete(&HFModelRepo{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
