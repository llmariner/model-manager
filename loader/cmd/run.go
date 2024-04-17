package main

import (
	"context"
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
	s := loader.New(
		st,
		c.BaseModels,
		filepath.Join(s3c.PathPrefix, s3c.BaseModelPathPrefix),
		newModelDownloader(c),
		newS3Client(c),
	)
	return s.Run(ctx, c.ModelLoadInterval)
}

func newModelDownloader(c *config.Config) loader.ModelDownloader {
	if c.Debug.Standalone {
		return &loader.NoopModelDownloader{}
	}
	return loader.NewHuggingFaceDownloader(c.Downloader.HuggingFace.CacheDir)
}

func newStore(c *config.Config) (*store.S, error) {
	if c.Debug.Standalone {
		dbInst, err := gorm.Open(sqlite.Open(c.Debug.SqlitePath), &gorm.Config{})
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
	return s3.NewClient(c.ObjectStore.S3)
}

func init() {
	runCmd.Flags().StringP(flagConfig, "c", "", "Configuration file path")
	_ = runCmd.MarkFlagRequired(flagConfig)
}
