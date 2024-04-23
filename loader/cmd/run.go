package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/llm-operator/model-manager/common/pkg/db"
	"github.com/llm-operator/model-manager/common/pkg/store"
	"github.com/llm-operator/model-manager/loader/internal/config"
	"github.com/llm-operator/model-manager/loader/internal/loader"
	"github.com/llm-operator/model-manager/loader/internal/s3"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const flagConfig = "config"

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := cmd.Flags().GetString(flagConfig)
		if err != nil {
			return err
		}

		c, err := config.Parse(path)
		if err != nil {
			return err
		}

		if err := c.Validate(); err != nil {
			return err
		}

		if err := run(cmd.Context(), &c); err != nil {
			return err
		}
		return nil
	},
}

func run(ctx context.Context, c *config.Config) error {
	st, err := newStore(c)
	if err != nil {
		return err
	}

	s3c := c.ObjectStore.S3
	d, err := newModelDownloader(c)
	if err != nil {
		return err
	}

	s := loader.New(
		st,
		c.BaseModels,
		filepath.Join(s3c.PathPrefix, s3c.BaseModelPathPrefix),
		d,
		newS3Client(c),
	)

	if c.RunOnce {
		return s.LoadBaseModels(ctx)
	}

	return s.Run(ctx, c.ModelLoadInterval)
}

func newModelDownloader(c *config.Config) (loader.ModelDownloader, error) {
	if c.Debug.Standalone {
		return &loader.NoopModelDownloader{}, nil
	}

	switch c.Downloader.Kind {
	case config.DownloaderKindS3:
		s3Client := s3.NewClient(s3.NewOptions{
			EndpointURL: c.Downloader.S3.EndpointURL,
			Region:      c.Downloader.S3.Region,
			Bucket:      c.Downloader.S3.Bucket,
			// Use anonymous credentials as the S3 bucket is public and we don't want to use the credential that is
			// used to upload the model.
			UseAnonymousCredentials: true,
		})
		return loader.NewS3Downloader(s3Client, c.Downloader.S3.PathPrefix), nil
	case config.DownloaderKindHuggingFace:
		return loader.NewHuggingFaceDownloader(c.Downloader.HuggingFace.CacheDir), nil
	default:
		return nil, fmt.Errorf("unknown downloader kind: %s", c.Downloader.Kind)
	}
}

func newStore(c *config.Config) (*store.S, error) {
	if c.Debug.Standalone || c.SkipDBUpdate {
		var path string
		if c.SkipDBUpdate {
			// Create an in-memory database so that writes to the database are not persisted.
			path = "file::memory:?cache=shared"
		} else {
			path = c.Debug.SqlitePath
		}
		dbInst, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		if err != nil {
			return nil, err
		}
		st := store.New(dbInst)
		if err := st.AutoMigrate(); err != nil {
			return nil, err
		}
		return st, nil
	}

	dbInst, err := db.OpenDB(c.Database)
	if err != nil {
		return nil, err
	}
	return store.New(dbInst), nil
}

func newS3Client(c *config.Config) loader.S3Client {
	if c.Debug.Standalone {
		return &loader.NoopS3Client{}
	}
	return s3.NewClient(s3.NewOptions{
		EndpointURL: c.ObjectStore.S3.EndpointURL,
		Region:      c.ObjectStore.S3.Region,
		Bucket:      c.ObjectStore.S3.Bucket,
	})

}

func init() {
	runCmd.Flags().StringP(flagConfig, "c", "", "Configuration file path")
	_ = runCmd.MarkFlagRequired(flagConfig)
}
