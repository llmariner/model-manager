package store

import (
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	v1 "github.com/llmariner/model-manager/api/v1"
)

// BaseModel represents a base model.
type BaseModel struct {
	gorm.Model

	TenantID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id"`

	ModelID string `gorm:"uniqueIndex:idx_base_model_model_id_tenant_id"`
	Path    string

	Formats []byte

	// GGUFModelPath is the path to the GGUF model.
	GGUFModelPath string
}

// UnmarshalModelFormats unmarshals model formats.
func UnmarshalModelFormats(b []byte) ([]v1.ModelFormat, error) {
	var formats v1.ModelFormats
	if err := proto.Unmarshal(b, &formats); err != nil {
		return nil, err
	}
	return formats.Formats, nil
}

// CreateBaseModel creates a model.
func (s *S) CreateBaseModel(
	modelID string,
	path string,
	formats []v1.ModelFormat,
	ggufModelPath string,
	tenantID string,
) (*BaseModel, error) {
	p := v1.ModelFormats{
		Formats: formats,
	}
	b, err := proto.Marshal(&p)
	if err != nil {
		return nil, err
	}

	m := &BaseModel{
		ModelID:       modelID,
		Path:          path,
		Formats:       b,
		GGUFModelPath: ggufModelPath,
		TenantID:      tenantID,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// DeleteBaseModel deletes a base model by model ID and tenant ID.
func (s *S) DeleteBaseModel(modelID, tenantID string) error {
	res := s.db.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&BaseModel{})
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// GetBaseModel returns a base model by model ID and tenant ID.
func (s *S) GetBaseModel(modelID, tenantID string) (*BaseModel, error) {
	var m BaseModel
	if err := s.db.Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Take(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListBaseModels returns all base models for a tenant.
func (s *S) ListBaseModels(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? ", tenantID).Order("id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}
