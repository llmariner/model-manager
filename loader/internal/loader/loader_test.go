package loader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	mv1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestLoadBaseModel(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{
			"dir0",
			"dir1",
			"dir0/dir2",
		},
		files: []string{
			"file0",
			"dir0/file1.gguf",
			"dir1/file2",
			"dir0/dir2/file3",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		[]string{"google/gemma-2b"},
		"models/base-models",
		downloader,
		s3Client,
		mc,
	)
	ld.tmpDir = "/tmp"
	err := ld.loadBaseModel(context.Background(), "google/gemma-2b")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/google/gemma-2b/dir0/dir2/file3",
		"models/base-models/google/gemma-2b/dir0/file1.gguf",
		"models/base-models/google/gemma-2b/dir1/file2",
		"models/base-models/google/gemma-2b/file0",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err := mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "google-gemma-2b",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_GGUF}, got.Formats)
	assert.Equal(t, "models/base-models/google/gemma-2b", got.Path)
	assert.Equal(t, "models/base-models/google/gemma-2b/dir0/file1.gguf", got.GgufModelPath)
}

func TestLoadBaseModel_HuggingFace(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{},
		files: []string{
			"config.json",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		[]string{"google/gemma-2b"},
		"models/base-models",
		downloader,
		s3Client,
		mc,
	)
	ld.tmpDir = "/tmp"
	err := ld.loadBaseModel(context.Background(), "google/gemma-2b")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/google/gemma-2b/config.json",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err := mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "google-gemma-2b",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_HUGGING_FACE}, got.Formats)
	assert.Equal(t, "models/base-models/google/gemma-2b", got.Path)
	assert.Empty(t, got.GgufModelPath)
}

type mockS3Client struct {
	uploadedKeys []string
}

func (c *mockS3Client) Upload(ctx context.Context, r io.Reader, key string) error {
	c.uploadedKeys = append(c.uploadedKeys, key)
	return nil
}

type fakeDownloader struct {
	dirs  []string
	files []string
}

func (d *fakeDownloader) download(ctx context.Context, modelName, desDir string) error {
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
