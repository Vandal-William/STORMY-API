package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the message service
type Config struct {
	Server    ServerConfig
	Cassandra CassandraConfig
	Redis     RedisConfig
	NATS      NATSConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// CassandraConfig holds Cassandra database configuration
type CassandraConfig struct {
	Hosts            []string
	Port             int
	Keyspace         string
	Username         string
	Password         string
	ConsistencyLevel string
	Timeout          time.Duration
	ConnectTimeout   time.Duration
	ReplicationFactor int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host string
	Port string
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "3001"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Cassandra: CassandraConfig{
			Hosts:             parseHosts(getEnv("CASSANDRA_HOSTS", "localhost")),
			Port:              getEnvInt("CASSANDRA_PORT", 9042),
			Keyspace:          getEnv("CASSANDRA_KEYSPACE", "message_service"),
			Username:          getEnv("CASSANDRA_USERNAME", ""),
			Password:          getEnv("CASSANDRA_PASSWORD", ""),
			ConsistencyLevel:  getEnv("CASSANDRA_CONSISTENCY", "LOCAL_ONE"),
			Timeout:           time.Duration(getEnvInt("CASSANDRA_TIMEOUT", 10)) * time.Second,
			ConnectTimeout:    time.Duration(getEnvInt("CASSANDRA_CONNECT_TIMEOUT", 10)) * time.Second,
			ReplicationFactor: getEnvInt("CASSANDRA_REPLICATION_FACTOR", 1),
		},
		Redis: RedisConfig{
			Host: getEnv("REDIS_HOST", "localhost"),
			Port: getEnv("REDIS_PORT", "6379"),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
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

// getEnvInt gets an environment variable as integer or returns a default value
func getEnvInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// parseHosts parses a comma-separated string of hosts
func parseHosts(hostsStr string) []string {
	if hostsStr == "" {
		return []string{"localhost"}
	}
	// For simple case, just split by comma
	hosts := []string{hostsStr}
	return hosts
}

// GetAddr returns the server address in format "host:port"
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}

// GetRedisAddr returns the Redis address
func (r *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// GetCassandraAddr returns the Cassandra address
func (c *CassandraConfig) GetAddr() string {
	if len(c.Hosts) > 0 {
		return c.Hosts[0]
	}
	return fmt.Sprintf("localhost:%d", c.Port)
}
