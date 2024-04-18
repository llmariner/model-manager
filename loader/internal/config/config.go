package config

import (
	"fmt"
	"os"
	"time"

	"github.com/llm-operator/model-manager/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// S3Config is the S3 configuration.
type S3Config struct {
	EndpointURL string `yaml:"endpointUrl"`
	Bucket      string `yaml:"bucket"`

	PathPrefix string `yaml:"pathPrefix"`

	// BaseModelPathPrefix is the path prefix for the base models in the object store. A model is stored under
	// <ObjectStore.S3.PathPrefix>/<BaseModelPathPrefix>.
	BaseModelPathPrefix string `yaml:"baseModelPathPrefix"`
}

// ObjectStoreConfig is the object store configuration.
type ObjectStoreConfig struct {
	S3 S3Config `yaml:"s3"`
}

// Validate validates the object store configuration.
func (c *ObjectStoreConfig) Validate() error {
	if c.S3.EndpointURL == "" {
		return fmt.Errorf("s3 endpoint url must be set")
	}
	if c.S3.Bucket == "" {
		return fmt.Errorf("s3 bucket must be set")
	}
	if c.S3.PathPrefix == "" {
		return fmt.Errorf("s3PathPrefix must be set")
	}
	if c.S3.BaseModelPathPrefix == "" {
		return fmt.Errorf("baseModelPathPrefix must be set")
	}
	return nil
}

// HuggingFaceDownloaderConfig is the Hugging Face downloader configuration.
type HuggingFaceDownloaderConfig struct {
	CacheDir string `yaml:"cacheDir"`
}

// DownloaderConfig is the downloader configuration.
type DownloaderConfig struct {
	HuggingFace HuggingFaceDownloaderConfig `yaml:"huggingFace"`
}

// Validate validates the downloader configuration.
func (c *DownloaderConfig) Validate() error {
	if c.HuggingFace.CacheDir == "" {
		return fmt.Errorf("cacheDir must be set")
	}

	return nil
}

// DebugConfig is the debug configuration.
type DebugConfig struct {
	Standalone bool   `yaml:"standalone"`
	SqlitePath string `yaml:"sqlitePath"`
}

// Config is the configuration.
type Config struct {
	Database db.Config `yaml:"database"`

	ObjectStore ObjectStoreConfig `yaml:"objectStore"`

	// BaseModels is the list of base models to load. Currently each model follows Hugging Face's model format.
	BaseModels []string `yaml:"baseModels"`

	ModelLoadInterval time.Duration `yaml:"modelLoadInterval"`

	Downloader DownloaderConfig `yaml:"downloader"`

	Debug DebugConfig `yaml:"debug"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if len(c.BaseModels) == 0 {
		return fmt.Errorf("baseModels must be set")
	}

	if c.ModelLoadInterval == 0 {
		return fmt.Errorf("modelloadInterval must be set")
	}

	if err := c.ObjectStore.Validate(); err != nil {
		return fmt.Errorf("objectStore: %s", err)
	}

	if err := c.Downloader.Validate(); err != nil {
		return fmt.Errorf("downloader: %s", err)

	}

	if c.Debug.Standalone {
		if c.Debug.SqlitePath == "" {
			return fmt.Errorf("sqlitePath must be set")
		}
	} else {
		if err := c.Database.Validate(); err != nil {
			return fmt.Errorf("database: %s", err)
		}
	}
	return nil
}

// Parse parses the configuration file at the given path, returning a new
// Config struct.
func Parse(path string) (Config, error) {
	var config Config

	b, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("config: read: %s", err)
	}

	if err = yaml.Unmarshal(b, &config); err != nil {
		return config, fmt.Errorf("config: unmarshal: %s", err)
	}
	return config, nil
}
