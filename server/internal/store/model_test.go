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
	_, err := st.GetPublishedModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateModel(ModelSpec{
		Key:         k,
		Path:        "path",
		IsPublished: true,
	})
	assert.NoError(t, err)

	gotM, err := st.GetPublishedModel(k)
	assert.NoError(t, err)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, tenantID, gotM.TenantID)

	gotMs, err := st.ListModelsByTenantID(tenantID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	k1 := ModelKey{
		ModelID:  "m1",
		TenantID: "tid1",
	}
	_, err = st.CreateModel(ModelSpec{
		Key:         k1,
		Path:        "path",
		IsPublished: true,
	})
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByTenantID(tenantID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	gotMs, err = st.ListAllPublishedModels()
	assert.NoError(t, err)
	assert.Len(t, gotMs, 2)

	err = st.DeleteModel(k)
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByTenantID(tenantID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)

	err = st.DeleteModel(k)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestModel_Unpublished(t *testing.T) {
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
	_, err := st.CreateModel(ModelSpec{
		Key:         k,
		Path:        "path",
		IsPublished: false,
	})
	assert.NoError(t, err)

	_, err = st.GetPublishedModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotMs, err := st.ListModelsByTenantID(tenantID, false)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	gotMs, err = st.ListModelsByTenantID(tenantID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)
}

func TestUpdateModel(t *testing.T) {
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
	err := st.UpdateModel(k, false)
	assert.Error(t, err)

	_, err = st.CreateModel(ModelSpec{
		Key:         k,
		Path:        "path",
		IsPublished: false,
	})
	assert.NoError(t, err)

	got, err := st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)

	err = st.UpdateModel(k, true)
	assert.NoError(t, err)
	got, err = st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.True(t, got.IsPublished)

	err = st.UpdateModel(k, false)
	assert.NoError(t, err)
	got, err = st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)
}
