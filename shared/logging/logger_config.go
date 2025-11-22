package logging

import (
	"go.uber.org/zap/zapcore"
)

type LoggerConfig struct {
	Level       zapcore.Level
	Encoding    string
	Development bool
	OutputPaths []string
}

type Option func(*LoggerConfig)

// WithLevel sets the logging level
func WithLevel(level zapcore.Level) Option {
	return func(c *LoggerConfig) {
		c.Level = level
	}
}

// WithEncoding sets the encoding format (json or console)
func WithEncoding(encoding string) Option {
	return func(c *LoggerConfig) {
		c.Encoding = encoding
	}
}

// WithDevelopment enables development mode
func WithDevelopment(dev bool) Option {
	return func(c *LoggerConfig) {
		c.Development = dev
	}
}

// WithOutputPaths sets output paths for logs
func WithOutputPaths(paths []string) Option {
	return func(c *LoggerConfig) {
		c.OutputPaths = paths
	}
}

// WithDebugLevel is a convenience function for debug level
func WithDebugLevel() Option {
	return WithLevel(zapcore.DebugLevel)
}

// WithInfoLevel is a convenience function for info level
func WithInfoLevel() Option {
	return WithLevel(zapcore.InfoLevel)
}

// WithWarnLevel is a convenience function for warn level
func WithWarnLevel() Option {
	return WithLevel(zapcore.WarnLevel)
}

// WithErrorLevel is a convenience function for error level
func WithErrorLevel() Option {
	return WithLevel(zapcore.ErrorLevel)
}

// WithJSONEncoding sets JSON encoding
func WithJSONEncoding() Option {
	return WithEncoding("json")
}

// WithConsoleEncoding sets console encoding
func WithConsoleEncoding() Option {
	return WithEncoding("console")
}
