package config

import (
	"github.com/spf13/viper"
)

// Log.Level constants
const (
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarning = "warning"
	LogLevelError   = "error"
)

// Log.Format constants
const (
	LogFormatJSON      = "json"
	LogFormatPlainText = "text"
)

// Config holds application configuration loaded from YAML.
type Config struct {
	Hosts struct {
		AutoCreate bool
		Folder     string
		Labels     map[string]string
	}
	API struct {
		Protocol string
		Host     string
		User     string
		Secret   string
	}
	Log struct {
		Level  string
		File   string
		Format string
		Source bool // Whether to include or not the source file on the log messages
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("Hosts.AutoCreate", true)
	v.SetDefault("Hosts.Folder", "/")
	v.SetDefault("API.Protocol", "https")
	v.SetDefault("API.Host", "localhost")
	v.SetDefault("Log.Level", LogLevelInfo)
	v.SetDefault("Log.File", "")
	v.SetDefault("Log.Format", LogFormatJSON)
	v.SetDefault("Log.SourceFile", false)
}

// Load reads configuration for the given site name and returns a Config
// populated with defaults and any values found in a YAML config file.
// If no config file is found, defaults are returned.
func Load(site string) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	v.AddConfigPath("/etc/cmk-piggyback-docker-resolver/")
	v.AddConfigPath(".")
	v.SetConfigType("yaml")
	v.SetConfigName(site)
	err := v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error and load defaults
		} else {
			return nil, err
		}
	}

	config := Config{}
	err = v.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, err
}
