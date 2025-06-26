package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Filters      FilterConfig       `mapstructure:"filters"`
	Location     LocationConfig     `mapstructure:"location"`
	Monitoring   MonitoringConfig   `mapstructure:"monitoring"`
	Notification NotificationConfig `mapstructure:"notification"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	URL      string `mapstructure:"url" validate:"required,url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// FilterConfig holds aircraft filtering configuration
type FilterConfig struct {
	AircraftType string `mapstructure:"aircraft_type"`
	MinAltitude  int    `mapstructure:"min_altitude"`
	MaxAltitude  int    `mapstructure:"max_altitude"`
	Military     bool   `mapstructure:"military"`
	Operator     string `mapstructure:"operator"`
	FlightNumber string `mapstructure:"flight_number"`
}

// LocationConfig holds location-related configuration
type LocationConfig struct {
	Latitude    float64 `mapstructure:"latitude"`
	Longitude   float64 `mapstructure:"longitude"`
	MaxDistance float64 `mapstructure:"max_distance"`
}

// MonitoringConfig holds monitoring-related configuration
type MonitoringConfig struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
	Debug        bool          `mapstructure:"debug"`
}

// NotificationConfig holds notification-related configuration
type NotificationConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Duration time.Duration `mapstructure:"duration"`
	// Tracking options
	NotifyOnCloserOnly bool          `mapstructure:"notify_on_closer_only"` // Only notify when aircraft get closer
	ReNotifyAfter      time.Duration `mapstructure:"re_notify_after"`       // Re-notify after this time even if not closer
	CleanupInterval    time.Duration `mapstructure:"cleanup_interval"`      // How often to clean up old aircraft history
}

// Load loads configuration from Viper (which is already set up by Cobra)
func Load(configFile string) (*Config, error) {
	// Set defaults
	setDefaults()

	// Set the config file
	viper.SetConfigFile(configFile)

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Read environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("GODAR")

	// Unmarshal the config from Viper
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.url", "")
	viper.SetDefault("server.username", "")
	viper.SetDefault("server.password", "")
	viper.SetDefault("filters.aircraft_type", "")
	viper.SetDefault("filters.min_altitude", 0)
	viper.SetDefault("filters.max_altitude", 0)
	viper.SetDefault("filters.military", false)
	viper.SetDefault("filters.operator", "")
	viper.SetDefault("filters.flight_number", "")
	viper.SetDefault("location.latitude", 0.0)
	viper.SetDefault("location.longitude", 0.0)
	viper.SetDefault("location.max_distance", 0.0)
	viper.SetDefault("monitoring.poll_interval", "60s")
	viper.SetDefault("monitoring.debug", false)
	viper.SetDefault("notification.enabled", false)
	viper.SetDefault("notification.duration", 30*time.Second)
	viper.SetDefault("notification.notify_on_closer_only", true)
	viper.SetDefault("notification.re_notify_after", 0*time.Second)
	viper.SetDefault("notification.cleanup_interval", 0*time.Second)
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Server.URL == "" {
		return fmt.Errorf("server URL is required")
	}

	if config.Filters.MinAltitude > 0 && config.Filters.MaxAltitude > 0 {
		if config.Filters.MinAltitude > config.Filters.MaxAltitude {
			return fmt.Errorf("min_altitude cannot be greater than max_altitude")
		}
	}

	if config.Location.Latitude != 0.0 || config.Location.Longitude != 0.0 {
		if config.Location.Latitude < -90 || config.Location.Latitude > 90 {
			return fmt.Errorf("latitude must be between -90 and 90")
		}
		if config.Location.Longitude < -180 || config.Location.Longitude > 180 {
			return fmt.Errorf("longitude must be between -180 and 180")
		}
	}

	if config.Monitoring.PollInterval < time.Second {
		return fmt.Errorf("poll_interval must be at least 1 second")
	}

	return nil
}

// SaveDefaultConfig saves a default configuration file
func SaveDefaultConfig(path string) error {
	setDefaults()
	return viper.WriteConfigAs(path)
}
