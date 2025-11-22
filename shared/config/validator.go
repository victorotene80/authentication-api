package config

import (
	"fmt"
	"strings"
)

func (c *Config) Validate() error {
	validators := []func() error{
		c.validateApp,
		c.validateServer,
		c.validateDatabase,
		c.validateJWT,
		c.validateSecurity,
	}

	for _, validator := range validators {
		if err := validator(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) validateApp() error {
	if c.App.Name == "" {
		return fmt.Errorf("app name cannot be empty")
	}

	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}

	if !validEnvs[c.App.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, or production)", c.App.Environment)
	}

	return nil
}

func (c *Config) validateServer() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1 and 65535)", c.Server.Port)
	}

	if c.Server.ReadTimeout < 0 {
		return fmt.Errorf("read timeout cannot be negative")
	}

	if c.Server.WriteTimeout < 0 {
		return fmt.Errorf("write timeout cannot be negative")
	}

	if c.Server.ShutdownTimeout < 0 {
		return fmt.Errorf("shutdown timeout cannot be negative")
	}

	return nil
}

func (c *Config) validateDatabase() error {
	if c.Database.Driver == "" {
		return fmt.Errorf("database driver cannot be empty")
	}

	validDrivers := map[string]bool{
		"postgres": true,
		"mysql":    true,
		"sqlite":   true,
	}

	if !validDrivers[c.Database.Driver] {
		return fmt.Errorf("unsupported database driver: %s", c.Database.Driver)
	}

	if c.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}

	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", c.Database.Port)
	}

	if c.Database.Name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	if c.Database.Username == "" {
		return fmt.Errorf("database username cannot be empty")
	}

	if c.Database.MaxOpenConns < 1 {
		return fmt.Errorf("max open connections must be at least 1")
	}

	if c.Database.MaxIdleConns < 1 {
		return fmt.Errorf("max idle connections must be at least 1")
	}

	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	return nil
}

func (c *Config) validateJWT() error {
	if c.JWT.AccessSecret == "" {
		return fmt.Errorf("JWT access secret cannot be empty")
	}

	if len(c.JWT.AccessSecret) < 32 {
    return fmt.Errorf("JWT access secret must be at least 32 characters long")
}

	if c.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT refresh secret cannot be empty")
	}

	if c.JWT.AccessSecret == c.JWT.RefreshSecret {
		return fmt.Errorf("JWT access and refresh secrets must be different")
	}

	if len(c.JWT.AccessSecret) < 32 {
		return fmt.Errorf("JWT access secret must be at least 32 characters long")
	}

	if len(c.JWT.RefreshSecret) < 32 {
		return fmt.Errorf("JWT refresh secret must be at least 32 characters long")
	}

	if c.JWT.AccessTokenDuration < 1 {
		return fmt.Errorf("access token duration must be positive")
	}

	if c.JWT.RefreshTokenDuration < 1 {
		return fmt.Errorf("refresh token duration must be positive")
	}

	if c.JWT.RefreshTokenDuration <= c.JWT.AccessTokenDuration {
		return fmt.Errorf("refresh token duration must be greater than access token duration")
	}

	if c.JWT.Issuer == "" {
		return fmt.Errorf("JWT issuer cannot be empty")
	}

	return nil
}

func (c *Config) validateSecurity() error {
	if c.Security.PasswordMinLength < 8 {
		return fmt.Errorf("password minimum length must be at least 8 characters")
	}

	if c.Security.MaxLoginAttempts < 1 {
		return fmt.Errorf("max login attempts must be at least 1")
	}

	if c.Security.LockoutDuration < 1 {
		return fmt.Errorf("lockout duration must be positive")
	}

	if c.Security.SessionTimeout < 1 {
		return fmt.Errorf("session timeout must be positive")
	}

	if c.Security.BcryptCost < 4 || c.Security.BcryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}

	return nil
}

// ValidateForEnvironment performs environment-specific validation
func (c *Config) ValidateForEnvironment() error {
	if c.App.IsProduction() {
		return c.validateProduction()
	}
	return nil
}

func (c *Config) validateProduction() error {
	var errors []string

	// Production-specific checks
	if c.App.Debug {
		errors = append(errors, "debug mode should be disabled in production")
	}

	if c.JWT.AccessSecret == "your-access-secret-key" || c.JWT.RefreshSecret == "your-refresh-secret-key" {
		errors = append(errors, "JWT secrets must be changed from default values in production")
	}

	if c.Database.SSLMode == "disable" {
		errors = append(errors, "SSL should be enabled for database connections in production")
	}

	if len(c.Server.CORS.AllowedOrigins) == 1 && c.Server.CORS.AllowedOrigins[0] == "*" {
		errors = append(errors, "CORS should not allow all origins in production")
	}

	if c.Security.BcryptCost < 10 {
		errors = append(errors, "bcrypt cost should be at least 10 in production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("production validation failed:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}