package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/llmariner/cluster-manager/pkg/status"
	v1 "github.com/llmariner/model-manager/api/v1"
	"gopkg.in/yaml.v3"
)

// AssumeRoleConfig is the assume role configuration.
type AssumeRoleConfig struct {
	RoleARN    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalId"`
}

func (c *AssumeRoleConfig) validate() error {
	if c.RoleARN == "" {
		return fmt.Errorf("roleArn must be set")
	}
	return nil
}

// S3Config is the S3 configuration.
type S3Config struct {
	EndpointURL string `yaml:"endpointUrl"`
	Region      string `yaml:"region"`

	Bucket string `yaml:"bucket"`

	PathPrefix string `yaml:"pathPrefix"`

	// BaseModelPathPrefix is the path prefix for the base models in the object store. A model is stored under
	// <ObjectStore.S3.PathPrefix>/<BaseModelPathPrefix>.
	BaseModelPathPrefix string `yaml:"baseModelPathPrefix"`

	AssumeRole *AssumeRoleConfig `yaml:"assumeRole"`
}

// ObjectStoreConfig is the object store configuration.
type ObjectStoreConfig struct {
	S3 S3Config `yaml:"s3"`
}

// validate validates the object store configuration.
func (c *ObjectStoreConfig) validate() error {
	if c.S3.Region == "" {
		return fmt.Errorf("s3 region must be set")
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
	if ar := c.S3.AssumeRole; ar != nil {
		if err := ar.validate(); err != nil {
			return fmt.Errorf("assumeRole: %s", err)
		}
	}
	return nil
}

// HuggingFaceDownloaderConfig is the Hugging Face downloader configuration.
type HuggingFaceDownloaderConfig struct {
	CacheDir string `yaml:"cacheDir"`
}

// S3DownloaderConfig is the S3 downloader configuration.
type S3DownloaderConfig struct {
	EndpointURL string `yaml:"endpointUrl"`
	Region      string `yaml:"region"`
	Bucket      string `yaml:"bucket"`
	PathPrefix  string `yaml:"pathPrefix"`
}

// OllamaDownloaderConfig is the Ollama downloader configuration.
type OllamaDownloaderConfig struct {
	Port int `yaml:"port"`
}

// DownloaderKind is the downloader kind.
type DownloaderKind string

const (
	// DownloaderKindS3 is the S3 downloader kind.
	DownloaderKindS3 DownloaderKind = "s3"
	// DownloaderKindHuggingFace is the Hugging Face downloader kind.
	DownloaderKindHuggingFace DownloaderKind = "huggingFace"
	// DownloaderKindOllama is the Ollama downloader kind.
	DownloaderKindOllama DownloaderKind = "ollama"
)

// DownloaderConfig is the downloader configuration.
type DownloaderConfig struct {
	Kind DownloaderKind `yaml:"kind"`

	HuggingFace HuggingFaceDownloaderConfig `yaml:"huggingFace"`
	S3          S3DownloaderConfig          `yaml:"s3"`
	Ollama      OllamaDownloaderConfig      `yaml:"ollama"`
}

// validate validates the downloader configuration.
func (c *DownloaderConfig) validate() error {
	switch c.Kind {
	case DownloaderKindS3:
		if c.S3.EndpointURL == "" {
			return fmt.Errorf("endpointUrl must be set")
		}
		if c.S3.Region == "" {
			return fmt.Errorf("region must be set")
		}
		if c.S3.Bucket == "" {
			return fmt.Errorf("bucket must be set")
		}
		if c.S3.PathPrefix == "" {
			return fmt.Errorf("pathPrefix must be set")
		}
	case DownloaderKindHuggingFace:
		if c.HuggingFace.CacheDir == "" {
			return fmt.Errorf("cacheDir must be set")
		}
	case DownloaderKindOllama:
		if c.Ollama.Port == 0 {
			return fmt.Errorf("port must be set")
		}
	default:
		return fmt.Errorf("unknown kind: %s", c.Kind)
	}

	return nil
}

// ModelConfig is the configuration for non-base-models.
type ModelConfig struct {
	Model            string `yaml:"model"`
	BaseModel        string `yaml:"baseModel"`
	AdapterType      string `yaml:"adapterType"`
	QuantizationType string `yaml:"quantizationType"`
}

// validate validates the model configuration.
func (c *ModelConfig) validate() error {
	if c.Model == "" {
		return fmt.Errorf("model must be set")
	}
	if c.BaseModel == "" {
		return fmt.Errorf("baseModel must be set")
	}
	if c.AdapterType != "" {
		a := ToAdapterType(c.AdapterType)
		if a != v1.AdapterType_ADAPTER_TYPE_LORA &&
			a != v1.AdapterType_ADAPTER_TYPE_QLORA {
			return fmt.Errorf("unsupported adapter type: %s", c.AdapterType)
		}
	}
	if c.QuantizationType != "" {
		q := ToQuantizationType(c.QuantizationType)
		if q != v1.QuantizationType_QUANTIZATION_TYPE_AWQ &&
			q != v1.QuantizationType_QUANTIZATION_TYPE_GGUF {
			return fmt.Errorf("unsupported quantization type: %s", c.QuantizationType)
		}
	}
	return nil
}

// DebugConfig is the debug configuration.
type DebugConfig struct {
	Standalone bool `yaml:"standalone"`
}

// WorkerTLSConfig is the worker TLS configuration.
type WorkerTLSConfig struct {
	Enable bool `yaml:"enable"`
}

// WorkerConfig is the worker configuration.
type WorkerConfig struct {
	TLS WorkerTLSConfig `yaml:"tls"`
}

// Config is the configuration.
type Config struct {
	ObjectStore ObjectStoreConfig `yaml:"objectStore"`

	// BaseModels is the list of base models to load. Currently each model follows Hugging Face's model format.
	BaseModels []string `yaml:"baseModels"`

	// Models is the list of fine-tuned or quantized models to load. Currently each model follows Hugging Face's model format.
	Models []ModelConfig `yaml:"models"`

	ModelLoadInterval time.Duration `yaml:"modelLoadInterval"`

	// RunOnce is set to true when models are loaded only once.
	RunOnce bool `yaml:"runOnce"`

	Downloader DownloaderConfig `yaml:"downloader"`

	ModelManagerServerWorkerServiceAddr string `yaml:"modelManagerServerWorkerServiceAddr"`

	ComponentStatusSender status.Config `yaml:"componentStatusSender"`

	Worker WorkerConfig `yaml:"worker"`

	Debug DebugConfig `yaml:"debug"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if len(c.BaseModels) == 0 {
		return fmt.Errorf("baseModels must be set")
	}

	for _, m := range c.Models {
		if err := m.validate(); err != nil {
			return fmt.Errorf("model: %s", err)
		}
	}

	if !c.RunOnce && c.ModelLoadInterval == 0 {
		return fmt.Errorf("modelloadInterval must be set")
	}

	if err := c.ObjectStore.validate(); err != nil {
		return fmt.Errorf("objectStore: %s", err)
	}

	if err := c.Downloader.validate(); err != nil {
		return fmt.Errorf("downloader: %s", err)
	}

	if err := c.ComponentStatusSender.Validate(); err != nil {
		return fmt.Errorf("componentStatusSender: %s", err)
	}

	if !c.Debug.Standalone {
		if c.ModelManagerServerWorkerServiceAddr == "" {
			return fmt.Errorf("model manager server worker service address must be set")
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

// ToAdapterType converts to v1.AdapterType.
func ToAdapterType(s string) v1.AdapterType {
	switch strings.ToLower(s) {
	case "lora":
		return v1.AdapterType_ADAPTER_TYPE_LORA
	case "qlora":
		return v1.AdapterType_ADAPTER_TYPE_QLORA
	default:
		return v1.AdapterType_ADAPTER_TYPE_UNSPECIFIED
	}
}

// ToQuantizationType converts to v1.QuantizationType.
func ToQuantizationType(s string) v1.QuantizationType {
	switch strings.ToLower(s) {
	case "awq":
		return v1.QuantizationType_QUANTIZATION_TYPE_AWQ
	case "gguf":
		return v1.QuantizationType_QUANTIZATION_TYPE_GGUF
	default:
		return v1.QuantizationType_QUANTIZATION_TYPE_UNSPECIFIED
	}
}
