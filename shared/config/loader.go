package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func Load() (*Config, error) {
	cfg := &Config{
		App:      loadAppConfig(),
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		Redis:    loadRedisConfig(),
		JWT:      loadJWTConfig(),
		Security: loadSecurityConfig(),
		Email:    loadEmailConfig(),
		SMS:      loadSMSConfig(),
		Logging:  loadLoggingConfig(),
		Tracing:  loadTracingConfig(),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func loadAppConfig() AppConfig {
	return AppConfig{
		Name:        getEnvOrDefault("APP_NAME", "authentication-service"),
		Version:     getEnvOrDefault("APP_VERSION", "1.0.0"),
		Environment: getEnvOrDefault("APP_ENV", "development"),
		Debug:       getEnvBool("APP_DEBUG", true),
	}
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Host:            getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		Port:            getEnvInt("SERVER_PORT", 8080),
		ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		CORS:            loadCORSConfig(),
	}
}

func loadCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods:   getEnvSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		AllowedHeaders:   getEnvSlice("CORS_ALLOWED_HEADERS", []string{"Accept", "Authorization", "Content-Type"}),
		ExposedHeaders:   getEnvSlice("CORS_EXPOSED_HEADERS", []string{""}),
		AllowCredentials: getEnvBool("CORS_ALLOW_CREDENTIALS", true),
		MaxAge:           getEnvInt("CORS_MAX_AGE", 300),
	}
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver:          getEnvOrDefault("DB_DRIVER", "postgres"),
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            getEnvInt("DB_PORT", 5432),
		Name:            getEnvOrDefault("DB_NAME", "auth_db"),
		Username:        getEnvOrDefault("DB_USERNAME", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", ""),
		SSLMode:         getEnvOrDefault("DB_SSL_MODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
	}
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:         getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:         getEnvInt("REDIS_PORT", 6379),
		Password:     getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:           getEnvInt("REDIS_DB", 0),
		PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
	}
}

func loadJWTConfig() JWTConfig {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	return JWTConfig{
		AccessSecret:         accessSecret,
		RefreshSecret:        refreshSecret,
		AccessTokenDuration:  getEnvDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		RefreshTokenDuration: getEnvDuration("JWT_REFRESH_DURATION", 7*24*time.Hour),
		Issuer:               getEnvOrDefault("JWT_ISSUER", "authentication-service"),
	}
}

func loadTracingConfig() TracerConfig {
	return TracerConfig{
		ServiceName:    getEnvOrDefault("APP_NAME", "authentication-service"),
		Environment:    getEnvOrDefault("APP_ENV", "development"),
		JaegerEndpoint: getEnvOrDefault("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		Enabled:        getEnvBool("TRACING_ENABLED", false),
	}
}

func loadSecurityConfig() SecurityConfig {
	return SecurityConfig{
		PasswordMinLength:      getEnvInt("SECURITY_PASSWORD_MIN_LENGTH", 8),
		PasswordRequireUpper:   getEnvBool("SECURITY_PASSWORD_REQUIRE_UPPER", true),
		PasswordRequireLower:   getEnvBool("SECURITY_PASSWORD_REQUIRE_LOWER", true),
		PasswordRequireNumber:  getEnvBool("SECURITY_PASSWORD_REQUIRE_NUMBER", true),
		PasswordRequireSpecial: getEnvBool("SECURITY_PASSWORD_REQUIRE_SPECIAL", true),
		MaxLoginAttempts:       getEnvInt("SECURITY_MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:        getEnvDuration("SECURITY_LOCKOUT_DURATION", 15*time.Minute),
		SessionTimeout:         getEnvDuration("SECURITY_SESSION_TIMEOUT", 24*time.Hour),
		BcryptCost:             getEnvInt("SECURITY_BCRYPT_COST", 12),
	}
}

func loadEmailConfig() EmailConfig {
	return EmailConfig{
		Provider: getEnvOrDefault("EMAIL_PROVIDER", "smtp"),
		Host:     getEnvOrDefault("EMAIL_HOST", "smtp.gmail.com"),
		Port:     getEnvInt("EMAIL_PORT", 587),
		Username: getEnvOrDefault("EMAIL_USERNAME", ""),
		Password: getEnvOrDefault("EMAIL_PASSWORD", ""),
		From:     getEnvOrDefault("EMAIL_FROM", "noreply@example.com"),
		FromName: getEnvOrDefault("EMAIL_FROM_NAME", "Authentication Service"),
	}
}

func loadSMSConfig() SMSConfig {
	return SMSConfig{
		Provider:  getEnvOrDefault("SMS_PROVIDER", "twilio"),
		AccountID: getEnvOrDefault("SMS_ACCOUNT_ID", ""),
		AuthToken: getEnvOrDefault("SMS_AUTH_TOKEN", ""),
		FromPhone: getEnvOrDefault("SMS_FROM_PHONE", ""),
	}
}

func loadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:      getEnvOrDefault("LOG_LEVEL", "info"),
		Format:     getEnvOrDefault("LOG_FORMAT", "json"),
		Output:     getEnvOrDefault("LOG_OUTPUT", "stdout"),
		FilePath:   getEnvOrDefault("LOG_FILE_PATH", "./logs/app.log"),
		MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
		MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
		Compress:   getEnvBool("LOG_COMPRESS", true),
	}
}



// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
