package metrics

import (
	"fmt"
	"net"
	"sync"
	"time"

	"authentication/shared/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogstashWriter writes logs to Logstash over TCP/UDP
type LogstashWriter struct {
	cfg        config.LoggingConfig
	conn       net.Conn
	mu         sync.RWMutex
	connected  bool
	logger     *zap.Logger // fallback logger
	stopCh     chan struct{}
	retryTimer *time.Timer
}

// NewLogstashWriter creates a new Logstash writer with configuration
func NewLogstashWriter(cfg config.LoggingConfig, fallbackLogger *zap.Logger) *LogstashWriter {
	return &LogstashWriter{
		cfg:    cfg,
		logger: fallbackLogger,
		stopCh: make(chan struct{}),
	}
}

// Start initiates connection and reconnection loop
func (w *LogstashWriter) Start() error {
	if !w.cfg.LogstashEnabled {
		return nil
	}

	// Initial connection attempt
	if err := w.connect(); err != nil {
		w.logger.Warn("initial Logstash connection failed, will retry in background",
			zap.Error(err),
			zap.String("addr", w.cfg.GetLogstashAddr()),
		)
	}

	// Start reconnection loop
	go w.reconnectionLoop()

	return nil
}

// reconnectionLoop continuously attempts to reconnect if disconnected
func (w *LogstashWriter) reconnectionLoop() {
	ticker := time.NewTicker(w.cfg.LogstashReconnectWait)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.mu.RLock()
			isConnected := w.connected
			w.mu.RUnlock()

			if !isConnected {
				if err := w.connect(); err != nil {
					w.logger.Debug("Logstash reconnection attempt failed",
						zap.Error(err),
						zap.String("addr", w.cfg.GetLogstashAddr()),
					)
				} else {
					w.logger.Info("Logstash connection restored",
						zap.String("addr", w.cfg.GetLogstashAddr()),
					)
				}
			}
		}
	}
}

// connect establishes connection to Logstash
func (w *LogstashWriter) connect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close existing connection if any
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
		w.connected = false
	}

	// Establish new connection
	dialer := net.Dialer{
		Timeout: w.cfg.LogstashTimeout,
	}

	conn, err := dialer.Dial(w.cfg.LogstashProtocol, w.cfg.GetLogstashAddr())
	if err != nil {
		return fmt.Errorf("failed to connect to Logstash: %w", err)
	}

	w.conn = conn
	w.connected = true

	w.logger.Info("connected to Logstash",
		zap.String("addr", w.cfg.GetLogstashAddr()),
		zap.String("protocol", w.cfg.LogstashProtocol),
	)

	return nil
}

// Write implements io.Writer interface
func (w *LogstashWriter) Write(p []byte) (n int, err error) {
	if !w.cfg.LogstashEnabled {
		return len(p), nil
	}

	w.mu.RLock()
	conn := w.conn
	connected := w.connected
	w.mu.RUnlock()

	if !connected || conn == nil {
		// Connection not available, drop the log (already logged to fallback)
		return 0, fmt.Errorf("not connected to Logstash")
	}

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(w.cfg.LogstashTimeout)); err != nil {
		w.markDisconnected()
		return 0, fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Write to Logstash
	n, err = conn.Write(p)
	if err != nil {
		w.markDisconnected()
		return n, fmt.Errorf("failed to write to Logstash: %w", err)
	}

	// Ensure newline for proper Logstash parsing
	if len(p) > 0 && p[len(p)-1] != '\n' {
		if _, err := conn.Write([]byte("\n")); err != nil {
			w.markDisconnected()
			return n, err
		}
	}

	return n, nil
}

// markDisconnected marks the connection as disconnected
func (w *LogstashWriter) markDisconnected() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.connected {
		w.connected = false
		if w.conn != nil {
			w.conn.Close()
			w.conn = nil
		}
		w.logger.Warn("Logstash connection lost, will attempt to reconnect",
			zap.String("addr", w.cfg.GetLogstashAddr()),
		)
	}
}

// Sync implements zapcore.WriteSyncer
func (w *LogstashWriter) Sync() error {
	// TCP connections don't need explicit sync
	return nil
}

// Close gracefully closes the Logstash writer
func (w *LogstashWriter) Close() error {
	close(w.stopCh)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		w.connected = false
		return err
	}

	return nil
}

// IsConnected returns whether the writer is currently connected
func (w *LogstashWriter) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// CreateLogstashCore creates a zapcore.Core that writes to Logstash
func CreateLogstashCore(cfg config.LoggingConfig, fallbackLogger *zap.Logger) (zapcore.Core, *LogstashWriter, error) {
	if !cfg.LogstashEnabled {
		return nil, nil, nil
	}

	writer := NewLogstashWriter(cfg, fallbackLogger)

	if err := writer.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start Logstash writer: %w", err)
	}

	// Create encoder for JSON format
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// Parse log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err == nil {
		// Use parsed level
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), level)

	return core, writer, nil
}
