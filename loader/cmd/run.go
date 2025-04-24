package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	cmstatus "github.com/llmariner/cluster-manager/pkg/status"
	laws "github.com/llmariner/common/pkg/aws"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/loader/internal/config"
	"github.com/llmariner/model-manager/loader/internal/loader"
	"github.com/llmariner/model-manager/loader/internal/s3"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func runCmd() *cobra.Command {
	var path string
	var logLevel int
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.Parse(path)
			if err != nil {
				return err
			}
			if err := c.Validate(); err != nil {
				return err
			}
			stdr.SetVerbosity(logLevel)
			if err := run(cmd.Context(), &c); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "config", "", "Path to the config file")
	cmd.Flags().IntVar(&logLevel, "v", 0, "Log level")
	_ = cmd.MarkFlagRequired("config")
	return cmd
}

func run(ctx context.Context, c *config.Config) error {
	logger := stdr.New(log.Default())
	ctx = logr.NewContext(ctx, logger)

	if err := auth.ValidateClusterRegistrationKey(); err != nil {
		return err
	}

	s3c := c.ObjectStore.S3

	var mclient loader.ModelClient
	if c.Debug.Standalone {
		mclient = loader.NewFakeModelClient()
	} else {
		conn, err := grpc.NewClient(c.ModelManagerServerWorkerServiceAddr, grpcOption(c))
		if err != nil {
			return err
		}
		mc := v1.NewModelsWorkerServiceClient(conn)
		if err := createStorageClass(ctx, mc, s3c.PathPrefix); err != nil {
			return err
		}
		mclient = mc
	}

	s3client, err := newS3Client(ctx, c)
	if err != nil {
		return err
	}
	sourceRepository := kindToSourceRepository(c.Downloader.Kind)
	if sourceRepository == v1.SourceRepository_SOURCE_REPOSITORY_UNSPECIFIED {
		return fmt.Errorf("invalid kind: %s", c.Downloader.Kind)
	}

	s := loader.New(
		s3c.Bucket,
		s3c.PathPrefix,
		s3c.BaseModelPathPrefix,
		&mdFactory{c: c},
		s3client,
		mclient,
		logger,
	)

	if c.RunOnce {
		return s.LoadModels(ctx, c.BaseModels, c.Models, sourceRepository)
	}

	ss, err := cmstatus.NewBeaconSender(c.ComponentStatusSender, grpcOption(c), logger)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error { return s.Run(ctx, c.BaseModels, c.Models, sourceRepository, c.ModelLoadInterval) })
	if c.ComponentStatusSender.Enable {
		eg.Go(func() error {
			ss.Run(ctx)
			return nil
		})
	}
	return eg.Wait()
}

func createStorageClass(ctx context.Context, mclient v1.ModelsWorkerServiceClient, pathPrefix string) error {
	ctx = auth.AppendWorkerAuthorization(ctx)

	_, err := mclient.GetStorageConfig(ctx, &v1.GetStorageConfigRequest{})
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok && s.Code() != codes.NotFound {
		return err
	}

	logr.FromContextOrDiscard(ctx).WithName("boot").Info("Creating a storage class", "pathPrefix", pathPrefix)
	_, err = mclient.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: pathPrefix,
	})
	return err
}

func kindToSourceRepository(kind config.DownloaderKind) v1.SourceRepository {
	switch kind {
	case config.DownloaderKindS3:
		return v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE
	case config.DownloaderKindHuggingFace:
		return v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE
	case config.DownloaderKindOllama:
		return v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA
	default:
		return v1.SourceRepository_SOURCE_REPOSITORY_UNSPECIFIED
	}
}

type mdFactory struct {
	c *config.Config
}

// Create creates a ModelDownloader.
func (f *mdFactory) Create(ctx context.Context, sourceRepository v1.SourceRepository) (loader.ModelDownloader, error) {
	logger := logr.FromContextOrDiscard(ctx)
	switch sourceRepository {
	case v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE:
		s3c := f.c.Downloader.S3
		opts := laws.NewS3ClientOptions{
			EndpointURL: s3c.EndpointURL,
			Region:      s3c.Region,
			// Use anonymous credentials when the S3 bucket is public and we don't want to use the credential that is
			// used to upload the model.
			UseAnonymousCredentials: s3c.IsPublic,
		}
		if ar := s3c.AssumeRole; ar != nil {
			opts.AssumeRole = &laws.AssumeRole{
				RoleARN:    ar.RoleARN,
				ExternalID: ar.ExternalID,
			}
		}
		s3Client, err := s3.NewClient(ctx, opts)
		if err != nil {
			return nil, err
		}
		return loader.NewS3Downloader(s3Client, s3c.Bucket, s3c.PathPrefix, logger), nil
	case v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE:
		return loader.NewHuggingFaceDownloader(f.c.Downloader.HuggingFace.CacheDir, logger), nil
	case v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA:
		return loader.NewOllamaDownloader(f.c.Downloader.Ollama.Port, logger), nil
	default:
		return nil, fmt.Errorf("unknown downloader source repository: %s", sourceRepository)
	}
}

func newS3Client(ctx context.Context, c *config.Config) (loader.S3Client, error) {
	s := c.ObjectStore.S3
	opts := laws.NewS3ClientOptions{
		EndpointURL: s.EndpointURL,
		Region:      s.Region,
	}
	if ar := s.AssumeRole; ar != nil {
		opts.AssumeRole = &laws.AssumeRole{
			RoleARN:    ar.RoleARN,
			ExternalID: ar.ExternalID,
		}
	}
	return s3.NewClient(ctx, opts)
}

func grpcOption(c *config.Config) grpc.DialOption {
	if c.Worker.TLS.Enable {
		return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}
