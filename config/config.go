package config

import (
	"github.com/spf13/viper"
)

// Config holds configuration
type Config struct {
	S3ApiServer *S3ApiServerConfig `yaml:"s3ApiServer"`
	Logging     *LoggingConfig     `yaml:"logging"`
}

// LoggingConfig holds configuration for logger
type LoggingConfig struct {
	Output string `yaml:"output"` // stdout, stderr or path to a file
	Level  string `yaml:"level"`  // DEBUG, INFO, WARN, ERROR, FATAL, PANIC
}

// HTTPConfig holds configuration for HTTP server
type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

// S3ApiServerConfig holds configuration for S3 api server
type S3ApiServerConfig struct {
	HTTP           *HTTPConfig `yaml:"http"`
	AdvertisedAddr string      `yaml:"advertisedAddr"`
	DataFolder     string      `yaml:"dataFolder"`
}

// ReadConfig reads configuration from viper
func ReadConfig() (*Config, error) {
	config := &Config{
		Logging: &LoggingConfig{
			Output: "stdout",
			Level:  "DEBUG",
		},
		S3ApiServer: &S3ApiServerConfig{
			HTTP: &HTTPConfig{
				Addr: ":8000",
			},
			AdvertisedAddr: "",
			DataFolder:     "/data/fakes3",
		},
	}
	err := viper.Unmarshal(config)
	return config, err
}
