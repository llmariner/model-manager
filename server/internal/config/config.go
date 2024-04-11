package config

import (
	"fmt"
	"os"

	"github.com/llm-operator/model-manager/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// S3Config is the S3 configuration.
type S3Config struct {
	PathPrefix string `yaml:"pathPrefix"`
}

// ObjectStoreConfig is the object store configuration.
type ObjectStoreConfig struct {
	S3 S3Config `yaml:"s3"`
}

// Validate validates the object store configuration.
func (c *ObjectStoreConfig) Validate() error {
	if c.S3.PathPrefix == "" {
		return fmt.Errorf("s3 path prefix must be set")
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
	HTTPPort         int `yaml:"httpPort"`
	GRPCPort         int `yaml:"grpcPort"`
	InternalGRPCPort int `yaml:"internalGrpcPort"`

	Database db.Config `yaml:"database"`

	ObjectStore ObjectStoreConfig `yaml:"objectStore"`

	Debug DebugConfig `yaml:"debug"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.HTTPPort <= 0 {
		return fmt.Errorf("httpPort must be greater than 0")
	}
	if c.GRPCPort <= 0 {
		return fmt.Errorf("grpcPort must be greater than 0")
	}
	if c.InternalGRPCPort <= 0 {
		return fmt.Errorf("internalGrpcPort must be greater than 0")
	}

	if err := c.ObjectStore.Validate(); err != nil {
		return fmt.Errorf("object store: %s", err)
	}

	if c.Debug.Standalone {
		if c.Debug.SqlitePath == "" {
			return fmt.Errorf("sqlite path must be set")
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
