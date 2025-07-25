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

	Name      string `gorm:"uniqueIndex:idx_hf_model_repo_name_tenant_id_project_id"`
	ModelID   string `gorm:"uniqueIndex:idx_hf_model_repo_model_id_tenant_id_project_id"`
	TenantID  string `gorm:"uniqueIndex:idx_hf_model_repo_name_tenant_id_project_id;uniqueIndex:idx_hf_model_repo_model_id_tenant_id_project_id"`
	ProjectID string `gorm:"uniqueIndex:idx_hf_model_repo_name_tenant_id_project_id;uniqueIndex:idx_hf_model_repo_model_id_tenant_id_project_id"`
}

// CreateHFModelRepo creates a model repo.
func (s *S) CreateHFModelRepo(
	name string,
	modelID string,
	tenantID string,
) (*HFModelRepo, error) {
	r := &HFModelRepo{
		Name:     name,
		ModelID:  modelID,
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

// DeleteHFModelRepoInTransactionByModelID deletes a model repo.
func DeleteHFModelRepoInTransactionByModelID(tx *gorm.DB, modelID, tenantID string) error {
	res := tx.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&HFModelRepo{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
