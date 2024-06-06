package store

import (
	"errors"
	"testing"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	_, err := st.GetBaseModel(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateBaseModel(modelID, "path", "gguf_model_path", tenantID)
	assert.NoError(t, err)

	gotM, err := st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, "path", gotM.Path)

	_, err = st.GetBaseModel(modelID, "different_tenant")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotMs, err := st.ListBaseModels(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.Equal(t, modelID, gotMs[0].ModelID)
	assert.Equal(t, "path", gotMs[0].Path)
	assert.Equal(t, "gguf_model_path", gotMs[0].GGUFModelPath)

	gotMs, err = st.ListBaseModels("different_tenant")
	assert.NoError(t, err)
	assert.Empty(t, gotMs)
}

func TestBaseModel_UniqueConstraint(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("m0", "path", "gguf_model_path", "t0")
	assert.NoError(t, err)

	_, err = st.CreateBaseModel("m0", "path", "gguf_model_path", "t0")
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	_, err = st.CreateBaseModel("m1", "path", "gguf_model_path", "t0")
	assert.NoError(t, err)

}
