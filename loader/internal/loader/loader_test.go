package loader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/llm-operator/model-manager/common/pkg/store"
	"github.com/stretchr/testify/assert"
)

func TestLoadBaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	downloader := &fakeDownloader{
		dirs: []string{
			"dir0",
			"dir1",
			"dir0/dir2",
		},
		files: []string{
			"file0",
			"dir0/file1",
			"dir1/file2",
			"dir0/dir2/file3",
		},
	}

	s3Client := &mockS3Client{}

	ld := New(
		st,
		[]string{"google/gemma-2b"},
		"models/base-models",
		downloader,
		s3Client,
	)
	err := ld.loadBaseModel(context.Background(), "google/gemma-2b")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/google/gemma-2b/dir0/dir2/file3",
		"models/base-models/google/gemma-2b/dir0/file1",
		"models/base-models/google/gemma-2b/dir1/file2",
		"models/base-models/google/gemma-2b/file0",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)
}

type mockS3Client struct {
	uploadedKeys []string
}

func (c *mockS3Client) Upload(r io.Reader, key string) error {
	c.uploadedKeys = append(c.uploadedKeys, key)
	return nil
}

type fakeDownloader struct {
	dirs  []string
	files []string
}

func (d *fakeDownloader) download(modelName, desDir string) error {
	for _, d := range d.dirs {
		if err := os.MkdirAll(filepath.Join(desDir, d), 0755); err != nil {
			return err
		}
	}
	for _, f := range d.files {
		file, err := os.Create(filepath.Join(desDir, f))
		if err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}
