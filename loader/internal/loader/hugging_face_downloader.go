package loader

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

// NewHuggingFaceDownloader creates a new HuggingFaceDownloader.
func NewHuggingFaceDownloader(cacheDir string) *HuggingFaceDownloader {
	return &HuggingFaceDownloader{
		cacheDir: cacheDir,
	}
}

// HuggingFaceDownloader downloads models from Hugging Face.
type HuggingFaceDownloader struct {
	cacheDir string
}

func (h *HuggingFaceDownloader) download(modelName, destDir string) error {
	// Download the image with huggingface-cli. If the access to the specified model is gated,
	// the environment variable HUGGING_FACE_HUB_TOKEN needs to be set.
	cmdline := []string{
		"huggingface-cli",
		"download",
		modelName,
		fmt.Sprintf("--local-dir=%s", destDir),
		fmt.Sprintf("--cache-dir=%s", h.cacheDir),
		"--local-dir-use-symlinks=True",
		"--quiet",
	}
	log.Printf("Downloading the image with command: %v\n", cmdline)
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to download the image: %s", errb.String())
		return fmt.Errorf("download: %s", err)
	}

	return nil
}
