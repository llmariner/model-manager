package store

import (
	"gorm.io/gorm"
)

// New creates a new store instance.
func New(db *gorm.DB) *S {
	return &S{
		db: db,
	}
}

// S represents the data store.
type S struct {
	db *gorm.DB
}

// Transaction runs a given function in a transaction.
func (s *S) Transaction(f func(*gorm.DB) error) error {
	return s.db.Transaction(f)
}

// AutoMigrate sets up the auto-migration task of the database.
func (s *S) AutoMigrate() error {
	return autoMigrate(s.db)
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&Model{},
		&BaseModel{},
		&HFModelRepo{},
		&ModelActivationStatus{},
		&StorageConfig{},
	); err != nil {
		return err
	}

	// The following indices are dropped with v1.22.0:
	indices := []struct {
		table interface{}
		name  string
	}{
		{&BaseModel{}, "idx_base_model_model_id_tenant_id"},
		{&ModelActivationStatus{}, "idx_model_activation_status_model_id_tenant_id"},
		{&HFModelRepo{}, "idx_hf_model_repo_model_id_tenant_id"},
		{&HFModelRepo{}, "idx_hf_model_repo_name_tenant_id"},
	}
	for _, idx := range indices {
		m := db.Migrator()
		if err := m.DropIndex(idx.table, idx.name); err != nil {
			// Ignore an error if its message is no such index:" + idx.name
			if err.Error() == "no such index: "+idx.name {
				continue
			}
		}
	}

	return nil
}
