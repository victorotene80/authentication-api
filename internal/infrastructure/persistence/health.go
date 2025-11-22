package persistence

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

type HealthChecker struct {
    db *sql.DB
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
    return &HealthChecker{db: db}
}

type HealthStatus struct {
    Status           string        `json:"status"`
    ResponseTime     time.Duration `json:"response_time_ms"`
    OpenConnections  int           `json:"open_connections"`
    InUse            int           `json:"in_use"`
    Idle             int           `json:"idle"`
    WaitCount        int64         `json:"wait_count"`
    WaitDuration     time.Duration `json:"wait_duration_ms"`
    MaxIdleClosed    int64         `json:"max_idle_closed"`
    MaxLifetimeClosed int64        `json:"max_lifetime_closed"`
}

func (h *HealthChecker) Check(ctx context.Context) (*HealthStatus, error) {
    start := time.Now()
    
    // Ping database
    if err := h.db.PingContext(ctx); err != nil {
        return &HealthStatus{
            Status:       "unhealthy",
            ResponseTime: time.Since(start),
        }, fmt.Errorf("database ping failed: %w", err)
    }
    
    responseTime := time.Since(start)
    
    // Get connection pool stats
    stats := h.db.Stats()
    
    return &HealthStatus{
        Status:            "healthy",
        ResponseTime:      responseTime,
        OpenConnections:   stats.OpenConnections,
        InUse:             stats.InUse,
        Idle:              stats.Idle,
        WaitCount:         stats.WaitCount,
        WaitDuration:      stats.WaitDuration,
        MaxIdleClosed:     stats.MaxIdleClosed,
        MaxLifetimeClosed: stats.MaxLifetimeClosed,
    }, nil
}

// CheckReadiness performs a more thorough check for readiness probe
func (h *HealthChecker) CheckReadiness(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    
    // Try to execute a simple query
    var result int
    err := h.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        return fmt.Errorf("readiness check failed: %w", err)
    }
    
    if result != 1 {
        return fmt.Errorf("unexpected query result: %d", result)
    }
    
    return nil
}