package loader

import (
	"context"
	"errors"
	"log"
	"path"
	"time"

	"github.com/llm-operator/model-manager/common/pkg/store"
	"gorm.io/gorm"
)

// New creates a new loader.
func New(store *store.S, baseModels []string, objectStorPathPrefix string) *L {
	return &L{
		store:                store,
		baseModels:           baseModels,
		objectStorPathPrefix: objectStorPathPrefix,
	}
}

// L is a loader.
type L struct {
	store *store.S

	baseModels []string

	// objectStorPathPrefix is the prefix of the path to the base models in the object stoer.
	objectStorPathPrefix string
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
	// Check if the model exists in the database.
	_, err := l.store.GetBaseModel(modelID)
	if err == nil {
		// Model exists. Do nothing.
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	log.Printf("Loading base model %q\n", modelID)

	mpath := path.Join(l.objectStorPathPrefix, modelID)

	if _, err := l.store.CreateBaseModel(modelID, mpath); err != nil {
		return err
	}

	log.Printf("Successfully loaded base model %q\n", modelID)
	return nil
}
