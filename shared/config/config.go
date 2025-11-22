// shared/config/config.go
package config

import (
	"fmt"
	"time"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Security SecurityConfig
	Email    EmailConfig
	SMS      SMSConfig
	Logging  LoggingConfig
	Tracing  TracerConfig
}

type TracerConfig struct {
	ServiceName    string
	Environment    string
	JaegerEndpoint string
	Enabled        bool
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Debug       bool
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Name            string
	Username        string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
}

type JWTConfig struct {
	AccessSecret         string
	RefreshSecret        string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type SecurityConfig struct {
	PasswordMinLength      int
	PasswordRequireUpper   bool
	PasswordRequireLower   bool
	PasswordRequireNumber  bool
	PasswordRequireSpecial bool
	MaxLoginAttempts       int
	LockoutDuration        time.Duration
	SessionTimeout         time.Duration
	BcryptCost             int
}

type EmailConfig struct {
	Provider string
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

type SMSConfig struct {
	Provider  string
	AccountID string
	AuthToken string
	FromPhone string
}

type LoggingConfig struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func (d DatabaseConfig) GetDSN() string {
	switch d.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			d.Host, d.Port, d.Username, d.Password, d.Name, d.SSLMode,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.Username, d.Password, d.Host, d.Port, d.Name,
		)
	default:
		return ""
	}
}

func (r RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func (s ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (a AppConfig) IsProduction() bool {
	return a.Environment == "production"
}

func (a AppConfig) IsDevelopment() bool {
	return a.Environment == "development"
}
