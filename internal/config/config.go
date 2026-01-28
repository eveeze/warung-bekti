package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Minio    MinioConfig
	JWT      JWTConfig
	App      AppConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host            string
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// MinioConfig holds Minio configuration
type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Environment string
	LogLevel    string
	APIVersion  string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "warung"),
			Password:        getEnv("DB_PASSWORD", "warung_secret"),
			DBName:          getEnv("DB_NAME", "warung_db"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		},
		Minio: MinioConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getBoolEnv("MINIO_USE_SSL", false),
			BucketName:      getEnv("MINIO_BUCKET_NAME", "warung-assets"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationHours: getIntEnv("JWT_EXPIRATION_HOURS", 24),
			Issuer:          getEnv("JWT_ISSUER", "warung-backend"),
		},
		App: AppConfig{
			Environment: getEnv("APP_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
			APIVersion:  getEnv("API_VERSION", "v1"),
		},
	}
}

// DSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

// Address returns the Redis address
func (c *RedisConfig) Address() string {
	return c.Host + ":" + c.Port
}

// ServerAddress returns the server address
func (c *ServerConfig) Address() string {
	return c.Host + ":" + c.Port
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
