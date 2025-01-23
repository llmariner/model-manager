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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		nil,
		nil,
		"models",
		"base-models",
		downloader,
		false,
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
		nil,
		nil,
		"models",
		"base-models",
		downloader,
		false,
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
		nil,
		nil,
		"models",
		"base-models",
		downloader,
		false,
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

func TestLoadBaseModel_MultipleGGUFFiles(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{},
		files: []string{
			"phi-4-Q3_K_L.gguf",
			"phi-4-Q3_K_M.gguf",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		nil,
		nil,
		"models",
		"base-models",
		downloader,
		true,
		s3Client,
		mc,
		testr.New(t),
	)
	ld.tmpDir = "/tmp"
	err := ld.loadBaseModel(context.Background(), "lmstudio-community/phi-4-GGUF")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L.gguf",
		"models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_M.gguf",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	// No model created for the HuggingFace repo name.
	ctx := context.Background()
	_, err = mc.GetBaseModelPath(ctx, &mv1.GetBaseModelPathRequest{
		Id: "lmstudio-community-phi-4-GGUF",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	for _, q := range []string{"K_L", "K_M"} {
		got, err := mc.GetBaseModelPath(ctx, &mv1.GetBaseModelPathRequest{
			Id: "lmstudio-community-phi-4-GGUF-phi-4-Q3_" + q,
		})
		assert.NoError(t, err)
		assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_GGUF}, got.Formats)
		assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_"+q, got.Path)
		assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_"+q+".gguf", got.GgufModelPath)
	}

	_, err = mc.GetHFModelRepo(ctx, &mv1.GetHFModelRepoRequest{Name: "lmstudio-community/phi-4-GGUF"})
	assert.NoError(t, err)
}

func TestLoadBaseModel_SelectedGGUFFile(t *testing.T) {
	downloader := &fakeDownloader{
		dirs: []string{},
		files: []string{
			"phi-4-Q3_K_M.gguf",
		},
	}

	s3Client := &mockS3Client{}
	mc := NewFakeModelClient()
	ld := New(
		nil,
		nil,
		"models",
		"base-models",
		downloader,
		true,
		s3Client,
		mc,
		testr.New(t),
	)
	ld.tmpDir = "/tmp"
	err := ld.loadBaseModel(context.Background(), "lmstudio-community/phi-4-GGUF/phi-4-Q3_K_M.gguf")
	assert.NoError(t, err)

	want := []string{
		"models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_M.gguf",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err := mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "lmstudio-community-phi-4-GGUF-phi-4-Q3_K_M.gguf",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_GGUF}, got.Formats)
	assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF", got.Path)
	assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_M.gguf", got.GgufModelPath)

	// Download another file.

	downloader.files = []string{
		"phi-4-Q3_K_L.gguf",
	}
	s3Client.uploadedKeys = []string{}
	err = ld.loadBaseModel(context.Background(), "lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L.gguf")
	assert.NoError(t, err)

	want = []string{
		"models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L.gguf",
	}
	assert.ElementsMatch(t, want, s3Client.uploadedKeys)

	got, err = mc.GetBaseModelPath(context.Background(), &mv1.GetBaseModelPathRequest{
		Id: "lmstudio-community-phi-4-GGUF-phi-4-Q3_K_L.gguf",
	})
	assert.NoError(t, err)
	assert.ElementsMatch(t, []mv1.ModelFormat{mv1.ModelFormat_MODEL_FORMAT_GGUF}, got.Formats)
	assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF", got.Path)
	assert.Equal(t, "models/base-models/lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L.gguf", got.GgufModelPath)
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
		false,
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

func TestBuildModelIDForGGUF(t *testing.T) {
	tcs := []struct {
		modelID           string
		ggufModelFilePath string
		want              string
	}{
		{
			modelID:           "lmstudio-community/phi-4-GGUF",
			ggufModelFilePath: "/tmp/phi-4-Q3_K_L.gguf",
			want:              "lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L",
		},
	}

	for _, tc := range tcs {
		got := buildModelIDForGGUF(tc.modelID, tc.ggufModelFilePath)
		assert.Equal(t, tc.want, got)
	}
}

func TestToLLMarinerModelID(t *testing.T) {
	tcs := []struct {
		id   string
		want string
	}{
		{
			id:   "meta-llama/Llama-3.3-70B-Instruct",
			want: "meta-llama-Llama-3.3-70B-Instruct",
		},
	}

	for _, tc := range tcs {
		got := toLLMarinerModelID(tc.id)
		assert.Equal(t, tc.want, got)
	}
}

func TestSplitHFRepoAndFile(t *testing.T) {
	tcs := []struct {
		modelID  string
		wantRepo string
		wantFile string
		wantErr  bool
	}{
		{
			modelID:  "lmstudio-community/phi-4-GGUF",
			wantRepo: "lmstudio-community/phi-4-GGUF",
			wantFile: "",
		},
		{
			modelID:  "lmstudio-community/phi-4-GGUF/phi-4-Q3_K_L.gguf",
			wantRepo: "lmstudio-community/phi-4-GGUF",
			wantFile: "phi-4-Q3_K_L.gguf",
		},
		{
			modelID: "a/b/c/d",
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.modelID, func(t *testing.T) {
			gotRepo, gotFile, err := splitHFRepoAndFile(tc.modelID)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tc.wantRepo, gotRepo)
			assert.Equal(t, tc.wantFile, gotFile)
		})
	}
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

func (d *fakeDownloader) download(ctx context.Context, modelName, filename string, desDir string) error {
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
