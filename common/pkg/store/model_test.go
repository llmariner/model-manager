package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "tid0"
	)

	k := ModelKey{
		ModelID:  modelID,
		TenantID: tenantID,
	}
	_, err := st.GetModel(k)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	err = st.CreateModel(k)
	assert.NoError(t, err)

	gotM, err := st.GetModel(k)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, tenantID, gotM.TenantID)

	gotMs, err := st.ListModelsByTenantID(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	k1 := ModelKey{
		ModelID:  "m1",
		TenantID: "tid1",
	}
	err = st.CreateModel(k1)
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByTenantID(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	err = st.DeleteModel(k)
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByTenantID(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)

	err = st.DeleteModel(k)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
