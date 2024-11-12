package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type SnapshotServerConfig struct {
	Address         string                   `mapstructure:"address"`
	Interval        string                   `mapstructure:"interval"`
	Endpoints       []SnapshotEndpointConfig `mapstructure:"endpoints"`
	Backend         *SnapshotBackendConfig   `mapstructure:"backend"`
	TimestampFormat string                   `mapstructure:"timestamp_format"`
}

type SnapshotEndpointConfig struct {
	Name     string `mapstructure:"name"`
	Endpoint string `mapstructure:"endpoint"`
}

type SnapshotBackendConfig struct {
	Type      string `mapstructure:"type"`
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"accesskey"`
	SecretKey string `mapstructure:"secretkey"`
}

func LoadConfig(path string) (*SnapshotServerConfig, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &SnapshotServerConfig{}

	if err := v.Unmarshal(cfg, viper.DecodeHook(func(src, dst reflect.Type, data interface{}) (interface{}, error) {
		if src.Kind() != reflect.String {
			return data, nil
		}

		str := data.(string)
		if strings.Contains(str, "${") || strings.HasPrefix(str, "$") {

			envVar := strings.Trim(str, "${}")
			envVar = strings.TrimPrefix(envVar, "$")

			if value, exists := os.LookupEnv(envVar); exists {
				return value, nil
			}
		}
		return data, nil
	})); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *SnapshotServerConfig) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		c.Address = ":8080"
	}

	if strings.TrimSpace(c.Interval) == "" {
		c.Interval = "60m"
	}

	if len(c.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint must be defined")
	}

	if c.Backend == nil {
		return fmt.Errorf("backend must be defined")
	}

	if strings.TrimSpace(c.TimestampFormat) == "" {
		c.TimestampFormat = "2006-01-02_15-04-05"
	}

	if strings.TrimSpace(c.Backend.Type) != "minio" {
		return fmt.Errorf("currently only 'minio' is as backend supported")
	}

	uniques := make(map[string]bool)
	for _, endpoint := range c.Endpoints {
		if uniques[endpoint.Name] {
			return fmt.Errorf("duplicate endpoint name found: %s", endpoint.Name)
		}

		if strings.TrimSpace(endpoint.Endpoint) == "" {
			return fmt.Errorf("entry '%s' must define an endpoint", endpoint.Name)
		}

		uniques[endpoint.Name] = true
	}

	return nil
}
