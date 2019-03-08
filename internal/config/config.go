package config

import (
	"time"

	"github.com/rs/zerolog"
	"go.soon.build/kit/config"
)

// application name
const APP_NAME = "sslcheck"

// Config stores configuration options set by configuration file or env vars
type Config struct {
	Log Log
	SSL SSLChecker `mapstructure:"ssl"`
}

// Log contains logging configuration
type Log struct {
	Console bool
	Verbose bool
	Level   string
}

// SSLChecker contains configuration for cert checker
type SSLChecker struct {
	ConnectTimeout   time.Duration
	WarnValidity     int
	CriticalValidity int
}

// Default is a default configuration setup with sane defaults
var Default = Config{
	Log{
		Level: zerolog.InfoLevel.String(),
	},
	SSLChecker{
		ConnectTimeout:   30 * time.Second,
		WarnValidity:     30,
		CriticalValidity: 14,
	},
}

// New constructs a new Config instance
func New(opts ...config.Option) (Config, error) {
	c := Default
	v := config.ViperWithDefaults("sslcheck")
	err := config.ReadInConfig(v, &c, opts...)
	if err != nil {
		return c, err
	}
	return c, nil
}
