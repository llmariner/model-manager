package loader

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	v1 "github.com/llmariner/model-manager/api/v1"
)

const defaultStatusSenderInterval = 10 * time.Second

type statusUpdateClient interface {
	updateStatus(ctx context.Context, msg string) error
}

func newBaseModelStatusUpdateClient(
	modelClient ModelClient,
	modelID string,
	projectID string,
) *baseModelStatusUpdateClient {
	return &baseModelStatusUpdateClient{
		modelClient: modelClient,
		modelID:     modelID,
		projectID:   projectID,
	}
}

type baseModelStatusUpdateClient struct {
	modelClient ModelClient
	modelID     string
	projectID   string
}

func (c *baseModelStatusUpdateClient) updateStatus(ctx context.Context, msg string) error {
	_, err := c.modelClient.UpdateBaseModelLoadingStatus(
		ctx,
		&v1.UpdateBaseModelLoadingStatusRequest{
			Id:            c.modelID,
			ProjectId:     c.projectID,
			StatusMessage: msg,
		},
	)
	return err
}

func newFineTunedModelStatusUpdateClient(
	modelClient ModelClient,
	modelID string,
) *fineTunedModelStatusUpdateClient {
	return &fineTunedModelStatusUpdateClient{
		modelClient: modelClient,
		modelID:     modelID,
	}
}

type fineTunedModelStatusUpdateClient struct {
	modelClient ModelClient
	modelID     string
}

func (c *fineTunedModelStatusUpdateClient) updateStatus(ctx context.Context, msg string) error {
	_, err := c.modelClient.UpdateModelLoadingStatus(
		ctx,
		&v1.UpdateModelLoadingStatusRequest{
			Id:            c.modelID,
			StatusMessage: msg,
		},
	)
	return err
}

// newStatusSender returns a new statusSender instance.
func newStatusSender(
	statusUpdateClient statusUpdateClient,
	downloadDir string,
) *statusSender {
	return &statusSender{
		statusUpdateClient: statusUpdateClient,
		downloadDir:        downloadDir,
	}
}

// statusSender is responsible for sending status updates on model file download and upload.
type statusSender struct {
	statusUpdateClient statusUpdateClient
	downloadDir        string

	numUploadedFiles int
	mu               sync.Mutex
}

func (s *statusSender) run(
	ctx context.Context,
	interval time.Duration,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(interval):
			if err := s.sendStatus(ctx); err != nil {
				return fmt.Errorf("failed to send status: %w", err)
			}
		}
	}
}

func (s *statusSender) setNumUploadedFiles(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.numUploadedFiles = n
}

func (s *statusSender) getNumUploadedFiles() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.numUploadedFiles
}

func (s *statusSender) getNumDownloadedFiles() (int, error) {
	var n int
	if err := filepath.Walk(s.downloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		n++
		return nil
	}); err != nil {
		return 0, fmt.Errorf("scan download directory: %s", err)
	}

	return n, nil
}

func (s *statusSender) sendStatus(ctx context.Context) error {
	up := s.getNumUploadedFiles()
	down, err := s.getNumDownloadedFiles()
	if err != nil {
		return fmt.Errorf("get number of downloaded files: %w", err)
	}

	statusMsg := fmt.Sprintf("downloaded files: %d, uploaded files: %d", down, up)
	if s.statusUpdateClient.updateStatus(ctx, statusMsg); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}
