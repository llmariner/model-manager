package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestHFModelRepo(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		repoName = "r0"
		tenantID = "t0"
	)

	_, err := st.GetHFModelRepo(repoName, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateHFModelRepo(repoName, tenantID)
	assert.NoError(t, err)

	gotR, err := st.GetHFModelRepo(repoName, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, repoName, gotR.Name)

	_, err = st.GetHFModelRepo(repoName, "different_tenant")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotRs, err := st.ListHFModelRepos(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotRs, 1)
	assert.Equal(t, repoName, gotRs[0].Name)

	gotRs, err = st.ListHFModelRepos("different_tenant")
	assert.NoError(t, err)
	assert.Empty(t, gotRs)
}
