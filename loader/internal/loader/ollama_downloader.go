package loader

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/go-logr/logr"
)

// NewOllamaDownloader creates a new OllamaDownloader.
func NewOllamaDownloader(port int, log logr.Logger) *OllamaDownloader {
	return &OllamaDownloader{
		log:  log.WithName("ollama"),
		port: port,
	}
}

// OllamaDownloader downloads models from Ollama.
type OllamaDownloader struct {
	port int
	log  logr.Logger
}

func (o *OllamaDownloader) serveServer(ctx context.Context) error {
	return o.runCommand(ctx, "serve")
}

func (o *OllamaDownloader) waitForReady(ctx context.Context) error {
	const (
		timeout = 30 * time.Second
		tick    = 1 * time.Second
	)
	o.log.Info("Waiting for Ollama to be ready")
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := o.runCommand(ctx, "list"); err == nil {
				return nil
			}
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for Ollama to be ready")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *OllamaDownloader) download(ctx context.Context, modelName, filename, destDir string) error {
	if err := os.Setenv("OLLAMA_HOST", fmt.Sprintf("0.0.0.0:%d", o.port)); err != nil {
		return err
	}
	if err := os.Setenv("OLLAMA_MODELS", destDir); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		if err := o.serveServer(ctx); err != nil {
			o.log.V(2).Info("stopped serving server", "reason", err)
			cancel()
		}
	}()
	if err := o.waitForReady(ctx); err != nil {
		return err
	}
	return o.runCommand(ctx, "pull", modelName)
}

func (o *OllamaDownloader) runCommand(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ollama", args...)
	var errb bytes.Buffer
	cmd.Stderr = &errb

	o.log.Info("Running Ollama command", "args", args)
	if err := cmd.Run(); err != nil {
		o.log.Error(err, "Failed to run ollama command", "args", args, "stderr", errb.String())
		return err
	}
	o.log.Info("Ollama command completed", "args", args)
	return nil
}
