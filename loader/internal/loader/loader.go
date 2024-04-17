package loader

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/llm-operator/model-manager/common/pkg/store"
	"gorm.io/gorm"
)

// ModelDownloader is an interface for downloading a model.
type ModelDownloader interface {
	download(modelName, destDir string) error
}

// NoopModelDownloader is a no-op model downloader.
type NoopModelDownloader struct {
}

func (d *NoopModelDownloader) download(modelName, destDir string) error {
	return nil
}

// New creates a new loader.
func New(
	store *store.S,
	baseModels []string,
	objectStorPathPrefix string,
	modelDownloader ModelDownloader,
) *L {
	return &L{
		store:                store,
		baseModels:           baseModels,
		objectStorPathPrefix: objectStorPathPrefix,
		modelDownloader:      modelDownloader,
	}
}

// L is a loader.
type L struct {
	store *store.S

	baseModels []string

	// objectStorPathPrefix is the prefix of the path to the base models in the object stoer.
	objectStorPathPrefix string

	modelDownloader ModelDownloader
}

// Run runs the loader.
func (l *L) Run(ctx context.Context, interval time.Duration) error {
	if err := l.loadBaseModels(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := l.loadBaseModels(ctx); err != nil {
				return err
			}
		}
	}
}

func (l *L) loadBaseModels(ctx context.Context) error {
	for _, baseModel := range l.baseModels {
		if err := l.loadBaseModel(ctx, baseModel); err != nil {
			return err
		}
	}
	return nil
}

func (l *L) loadBaseModel(ctx context.Context, modelID string) error {
	// First check if the model exists in the database.
	_, err := l.store.GetBaseModel(modelID)
	if err == nil {
		// Model exists. Do nothing.
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	log.Printf("Started loading base model %q\n", modelID)

	dir, err := os.MkdirTemp("/tmp", "base-model")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	log.Printf("Downloading base model %q\n", modelID)
	if err := l.modelDownloader.download(modelID, dir); err != nil {
		return err
	}

	log.Printf("Uploading base model %q to the object store\n", modelID)

	// TODO(kenji): Upload all the files under the dir.

	mpath := filepath.Join(l.objectStorPathPrefix, modelID)

	if _, err := l.store.CreateBaseModel(modelID, mpath); err != nil {
		return err
	}

	log.Printf("Successfully loaded base model %q\n", modelID)
	return nil
}
