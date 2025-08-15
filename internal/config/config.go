package config

import (
	"github.com/spf13/viper"
)

// Config holds server configuration
type Config struct {
	Specs    []string `mapstructure:"specs"`
	Port     int      `mapstructure:"port"`
	Mode     string   `mapstructure:"mode"`
	LogLevel string   `mapstructure:"log_level"`
}

// Load loads configuration from file
func Load(configFile string) (*Config, error) {
	viper.SetDefault("port", 3000)
	viper.SetDefault("mode", "stdio")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("specs", []string{})

	viper.SetEnvPrefix("MCP")
	viper.AutomaticEnv()

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
