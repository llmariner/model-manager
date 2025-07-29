package loader

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusSender(t *testing.T) {
	client := &fakeStatusUpdateClient{}

	tmpDir, err := os.MkdirTemp("/tmp", "status_sender_test")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Create a test model file.
	f, err := os.Create(tmpDir + "/modelfile.txt")
	assert.NoError(t, err)
	_, err = f.WriteString("This is a test model file.")
	assert.NoError(t, err)
	err = f.Close()
	assert.NoError(t, err)

	s := newStatusSender(client, tmpDir)
	s.setNumUploadedFiles(2)
	err = s.sendStatus(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, "downloaded files: 1, uploaded files: 2", client.capturedMsg)
}

type fakeStatusUpdateClient struct {
	capturedMsg string
}

func (c *fakeStatusUpdateClient) updateStatus(ctx context.Context, msg string) error {
	c.capturedMsg = msg
	return nil
}
