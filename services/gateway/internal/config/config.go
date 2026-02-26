package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for the gateway service
type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// ServicesConfig holds downstream services URLs
type ServicesConfig struct {
	UserURL         string
	MessageURL      string
	PresenceURL     string
	NotificationURL string
	ModerationURL   string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Services: ServicesConfig{
			UserURL:         os.Getenv("USER_SERVICE_URL"),
			MessageURL:      os.Getenv("MESSAGE_SERVICE_URL"),
			PresenceURL:     os.Getenv("PRESENCE_SERVICE_URL"),
			NotificationURL: os.Getenv("NOTIFICATION_SERVICE_URL"),
			ModerationURL:   os.Getenv("MODERATION_SERVICE_URL"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// GetAddr returns the server address in format "host:port"
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}
