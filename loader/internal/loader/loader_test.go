package loader

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr/testr"
	mv1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/loader/internal/config"
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
		nil,
		"models",
		"base-models",
		downloader,
		s3Client,
		mc,
		testr.New(t),
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
		nil,
		"models",
		"base-models",
		downloader,
		s3Client,
		mc,
		testr.New(t),
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

func TestLoadBaseModel_NvidiaTriton(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{
			"repo",
			"repo/llama3",
			"repo/llama3/tensorrt_llm",
		},
		files: []string{
			"repo/llama3/tensorrt_llm/config.pbtxt",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		[]string{"meta-llama/Meta-Llama-3.1-70B-Instruct-awq-triton"},
		nil,
		"models",
		"base-models",
		downloader,
		s3Client,
		mc,
		testr.New(t),
	)
	ld.tmpDir = "/tmp"
	err := ld.loadBaseModel(context.Background(), "meta-llama/Meta-Llama-3.1-70B-Instruct-awq-triton")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/meta-llama/Meta-Llama-3.1-70B-Instruct-awq-triton/repo/llama3/tensorrt_llm/config.pbtxt",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err := mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "meta-llama-Meta-Llama-3.1-70B-Instruct-awq-triton",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_NVIDIA_TRITON}, got.Formats)
	assert.Equal(t, "models/base-models/meta-llama/Meta-Llama-3.1-70B-Instruct-awq-triton", got.Path)
	assert.Empty(t, got.GgufModelPath)
}

func TestLoadModel_HuggingFace(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{},
		files: []string{
			"adapter_config.json",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		nil,
		[]config.ModelConfig{
			{
				Model:       "abc/lora1",
				BaseModel:   "google/gemma-2b",
				AdapterType: "lora",
			},
		},
		"models",
		"base-models",
		downloader,
		s3Client,
		mc,
		testr.New(t),
	)
	ld.tmpDir = "/tmp"
	err := ld.loadModel(context.Background(), ld.models[0])
	assert.NoError(t, err)

	want := []string{
		"models/base-models/google/gemma-2b/adapter_config.json",
		"models/default-tenant-id/abc/lora1/adapter_config.json",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err := mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "google-gemma-2b",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_HUGGING_FACE}, got.Formats)
	assert.Equal(t, "models/base-models/google/gemma-2b", got.Path)
	assert.Empty(t, got.GgufModelPath)

	ret, err := mc.GetModelPath(context.Background(), &mv1.GetModelPathRequest{
		Id: "abc-lora1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "models/default-tenant-id/abc/lora1", ret.Path)
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
