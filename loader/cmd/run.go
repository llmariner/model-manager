package main

import (
	"context"
	"fmt"
	"path/filepath"

	mv1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/loader/internal/config"
	"github.com/llm-operator/model-manager/loader/internal/loader"
	"github.com/llm-operator/model-manager/loader/internal/s3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	s3c := c.ObjectStore.S3
	d, err := newModelDownloader(c)
	if err != nil {
		return err
	}

	var mclient loader.ModelClient
	if c.Debug.Standalone {
		mclient = loader.NewFakeModelClient()
	} else {
		conn, err := grpc.Dial(c.ModelManagerWorkerServiceServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		mclient = mv1.NewModelsWorkerServiceClient(conn)
	}

	s := loader.New(
		c.BaseModels,
		filepath.Join(s3c.PathPrefix, s3c.BaseModelPathPrefix),
		d,
		newS3Client(c),
		mclient,
	)

	if c.RunOnce {
		return s.LoadBaseModels(ctx)
	}

	return s.Run(ctx, c.ModelLoadInterval)
}

func newModelDownloader(c *config.Config) (loader.ModelDownloader, error) {
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

func newS3Client(c *config.Config) loader.S3Client {
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
