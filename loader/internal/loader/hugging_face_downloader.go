package loader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/go-logr/logr"
)

// NewHuggingFaceDownloader creates a new HuggingFaceDownloader.
func NewHuggingFaceDownloader(cacheDir string, log logr.Logger) *HuggingFaceDownloader {
	return &HuggingFaceDownloader{
		cacheDir: cacheDir,
		log:      log.WithName("huggingface"),
	}
}

// HuggingFaceDownloader downloads models from Hugging Face.
type HuggingFaceDownloader struct {
	cacheDir string
	log      logr.Logger
}

func (h *HuggingFaceDownloader) download(ctx context.Context, modelName, filename, destDir string) error {
	// Download the image with huggingface-cli. If the access to the specified model is gated,
	// the environment variable HUGGING_FACE_HUB_TOKEN needs to be set.
	cmdline := []string{
		"huggingface-cli",
		"download",
		modelName,
	}
	if filename != "" {
		cmdline = append(cmdline, filename)
	}

	cmdline = append(cmdline,
		fmt.Sprintf("--local-dir=%s", destDir),
		fmt.Sprintf("--cache-dir=%s", h.cacheDir),
		"--local-dir-use-symlinks=True",
		"--quiet",
	)
	h.log.Info("Downloading the image with command", "cmd", cmdline)
	cmd := exec.CommandContext(ctx, cmdline[0], cmdline[1:]...)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		h.log.Error(errors.New(errb.String()), "Failed to download the image")
		return fmt.Errorf("download: %s", err)
	}
	return nil
}
