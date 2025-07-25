package store

import "gorm.io/gorm"

// ModelKey is a unique key for base models and fine-tuned models.
type ModelKey struct {
	ModelID   string
	TenantID  string
	ProjectID string
}

func (k *ModelKey) buildQuery(tx *gorm.DB) *gorm.DB {
	tx = tx.Where("model_id = ?", k.ModelID).
		Where("tenant_id = ?", k.TenantID)
	if k.ProjectID != "" {
		tx = tx.Where("project_id = ?", k.ProjectID)
	} else {
		// For non-project scoped resources. Check 'IS NULL' for backward compatibility.
		tx = tx.Where("(project_id IS NULL OR project_id = '')")
	}
	return tx
}
