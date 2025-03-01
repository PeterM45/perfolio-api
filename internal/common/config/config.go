package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port         int           `mapstructure:"port"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	} `mapstructure:"server"`

	Database struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		Name            string        `mapstructure:"name"`
		SSLMode         string        `mapstructure:"ssl_mode"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"database"`

	Auth struct {
		JWTSecret      string        `mapstructure:"jwt_secret"`
		TokenExpiry    time.Duration `mapstructure:"token_expiry"`
		ClerkSecretKey string        `mapstructure:"clerk_secret_key"`
	} `mapstructure:"auth"`

	Cache struct {
		Type       string        `mapstructure:"type"`
		RedisURL   string        `mapstructure:"redis_url"`
		DefaultTTL time.Duration `mapstructure:"default_ttl"`
	} `mapstructure:"cache"`

	LogLevel string `mapstructure:"log_level"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	viper.AutomaticEnv()

	// Default values
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", time.Second*10)
	viper.SetDefault("server.write_timeout", time.Second*10)
	viper.SetDefault("server.idle_timeout", time.Second*60)

	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", time.Minute*5)

	viper.SetDefault("cache.type", "memory")
	viper.SetDefault("cache.default_ttl", time.Minute*5)

	viper.SetDefault("log_level", "info")

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
