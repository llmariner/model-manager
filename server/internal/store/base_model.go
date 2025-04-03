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

	SourceRepository v1.SourceRepository
	LoadingStatus    v1.ModelLoadingStatus

	LoadingFailureReason string
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
	sourceRepository v1.SourceRepository,
	tenantID string,
) (*BaseModel, error) {
	b, err := marshalFormats(formats)
	if err != nil {
		return nil, err
	}

	m := &BaseModel{
		ModelID:          modelID,
		Path:             path,
		Formats:          b,
		GGUFModelPath:    ggufModelPath,
		LoadingStatus:    v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		SourceRepository: sourceRepository,
		TenantID:         tenantID,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// CreateBaseModelWithLoadingRequested creates a model with the requested loading status.
func (s *S) CreateBaseModelWithLoadingRequested(
	modelID string,
	sourceRepository v1.SourceRepository,
	tenantID string,
) (*BaseModel, error) {
	m := &BaseModel{
		ModelID:          modelID,
		SourceRepository: sourceRepository,
		LoadingStatus:    v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
		TenantID:         tenantID,
	}
	if err := s.db.Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

// DeleteBaseModel deletes a base model by model ID and tenant ID.
func (s *S) DeleteBaseModel(modelID, tenantID string) error {
	return DeleteBaseModelInTransaction(s.db, modelID, tenantID)
}

// DeleteBaseModelInTransaction deletes a base model by model ID and tenant ID.
func DeleteBaseModelInTransaction(tx *gorm.DB, modelID, tenantID string) error {
	res := tx.Unscoped().Where("model_id = ? AND tenant_id = ?", modelID, tenantID).Delete(&BaseModel{})
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
	if err := s.db.Where("tenant_id = ? ", tenantID).Order("model_id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// ListBaseModelsWithPagination finds base models with pagination. Models are returned with a descending order of ID.
func (s *S) ListBaseModelsWithPagination(tenantID string, afterID uint, limit int) ([]*BaseModel, bool, error) {
	var ms []*BaseModel
	q := s.db.Where("tenant_id = ?", tenantID)
	if afterID > 0 {
		q = q.Where("id < ?", afterID)
	}
	if err := q.Order("id DESC").Limit(limit + 1).Find(&ms).Error; err != nil {
		return nil, false, err
	}

	var hasMore bool
	if len(ms) > limit {
		ms = ms[:limit]
		hasMore = true
	}
	return ms, hasMore, nil
}

// ListUnloadedBaseModels returns all unloaded base models with the requested loading status.
func (s *S) ListUnloadedBaseModels(tenantID string) ([]*BaseModel, error) {
	var ms []*BaseModel
	if err := s.db.Where("tenant_id = ? AND loading_status = ?", tenantID, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED).
		Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	return ms, nil
}

// updateBaseModel updates the model if the current status matches with the given one.
func (s *S) updateBaseModel(
	modelID string,
	tenantID string,
	curr v1.ModelLoadingStatus,
	updates map[string]interface{},
) error {
	res := s.db.Model(&BaseModel{}).
		Where("model_id = ? AND tenant_id = ? AND loading_status = ?", modelID, tenantID, curr).
		Updates(updates)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

// UpdateBaseModelToLoadingStatus updates the loading status to LOADING.
func (s *S) UpdateBaseModelToLoadingStatus(modelID string, tenantID string) error {
	return s.updateBaseModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
		map[string]interface{}{
			"loading_status": v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		},
	)
}

// UpdateBaseModelToSucceededStatus updates the loading status to SUCCEEDED and updates other relevant information.
func (s *S) UpdateBaseModelToSucceededStatus(
	modelID string,
	tenantID string,
	path string,
	formats []v1.ModelFormat,
	ggufModelPath string,
) error {
	b, err := marshalFormats(formats)
	if err != nil {
		return err
	}

	return s.updateBaseModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"path":            path,
			"formats":         b,
			"gguf_model_path": ggufModelPath,
			"loading_status":  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		},
	)
}

// UpdateBaseModelToFailedStatus updates the loading status to FAILED and updates other relevant information.
func (s *S) UpdateBaseModelToFailedStatus(
	modelID string,
	tenantID string,
	failureReason string,
) error {
	return s.updateBaseModel(
		modelID,
		tenantID,
		v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
		map[string]interface{}{
			"loading_failure_reason": failureReason,
			"loading_status":         v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED,
		},
	)
}

// CountBaseModels counts the total number of base models.
func (s *S) CountBaseModels(tenantID string) (int64, error) {
	var count int64
	if err := s.db.Model(&BaseModel{}).Where("tenant_id = ? ", tenantID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func marshalFormats(formats []v1.ModelFormat) ([]byte, error) {
	p := v1.ModelFormats{
		Formats: formats,
	}
	return proto.Marshal(&p)
}
