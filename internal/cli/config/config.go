// Package config wraps Viper to load DotnetArchitect's configuration from
// (in increasing priority order) defaults, a config file, environment
// variables, and CLI flags. Only the cli package should import this
// package; business logic packages must not know Viper exists.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Defaults for options that aren't provided anywhere else.
const (
	DefaultPresentation = "WebApi"
	DefaultDatabase     = "PostgreSQL"
)

// Config wraps a Viper instance configured for DotnetArchitect.
type Config struct {
	v *viper.Viper
}

// Load builds a Config, searching for a "dotnet-architect.yaml" /
// ".dotnet-architect.yaml" config file in the current directory and the
// user's home directory, and binding the DOTNET_ARCHITECT_ environment
// variable prefix. Missing config files are not an error - all settings
// have sane defaults or are supplied via flags.
func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigName("dotnet-architect")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME")

	v.SetEnvPrefix("DOTNET_ARCHITECT")
	v.AutomaticEnv()

	v.SetDefault("presentation", DefaultPresentation)
	v.SetDefault("database", DefaultDatabase)
	v.SetDefault("build", true)
	v.SetDefault("createProjectFolder", true)

	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); !notFound {
			return nil, fmt.Errorf("config: failed to read config file: %w", err)
		}
	}

	return &Config{v: v}, nil
}

// String returns a string setting.
func (c *Config) String(key string) string {
	return c.v.GetString(key)
}

// Bool returns a bool setting.
func (c *Config) Bool(key string) bool {
	return c.v.GetBool(key)
}
