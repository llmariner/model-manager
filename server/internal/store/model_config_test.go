package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModelConfig(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	c := &ModelConfig{
		ModelID:       "m0",
		ProjectID:     "p0",
		TenantID:      "t0",
		EncodedConfig: []byte("test"),
	}

	err := st.CreateModelConfig(c)
	assert.NoError(t, err)

	k := ModelKey{
		ModelID:   c.ModelID,
		ProjectID: c.ProjectID,
		TenantID:  c.TenantID,
	}

	got, err := st.GetModelConfig(k)
	assert.NoError(t, err)
	assert.Equal(t, c.EncodedConfig, got.EncodedConfig)

	err = st.UpdateModelConfig(k, []byte("updated"))
	assert.NoError(t, err)

	got, err = st.GetModelConfig(k)
	assert.NoError(t, err)
	assert.Equal(t, []byte("updated"), got.EncodedConfig)

	err = st.DeleteModelConfig(k)
	assert.NoError(t, err)

	_, err = st.GetModelConfig(k)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
