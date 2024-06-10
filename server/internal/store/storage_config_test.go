package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestStorageConfig(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		tenantID = "tid0"
	)

	_, err := st.GetStorageConfig(tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	c, err := st.CreateStorageConfig(tenantID, "path")
	assert.NoError(t, err)
	assert.Equal(t, "path", c.PathPrefix)

	c, err = st.GetStorageConfig(tenantID)
	assert.NoError(t, err)
	assert.Equal(t, "path", c.PathPrefix)

	_, err = st.GetStorageConfig("different-tenant")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}
