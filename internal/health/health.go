package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"connectrpc.com/grpchealth"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
)

// Status represents the health status of a service or dependency.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

// CheckResult represents the health check result for a single dependency.
type CheckResult struct {
	Status    Status `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

// HealthStatus represents the overall health status of the service.
type HealthStatus struct {
	Status  Status                 `json:"status"`
	Version string                 `json:"version,omitempty"`
	Checks  map[string]CheckResult `json:"checks,omitempty"`
}

// Checker performs health checks on service dependencies.
type Checker struct {
	db      *sql.DB
	natsURL string
	version string
}

// NewChecker creates a new health checker with the given dependencies.
func NewChecker(db *sql.DB, natsURL, version string) *Checker {
	return &Checker{
		db:      db,
		natsURL: natsURL,
		version: version,
	}
}

// Check performs health checks on all dependencies and returns the overall status.
func (c *Checker) Check(ctx context.Context) *HealthStatus {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := &HealthStatus{
		Status:  StatusHealthy,
		Version: c.version,
		Checks:  make(map[string]CheckResult),
	}

	// PostgreSQL check
	if c.db != nil {
		start := time.Now()
		if err := c.db.PingContext(checkCtx); err != nil {
			status.Status = StatusUnhealthy
			status.Checks["postgres"] = CheckResult{
				Status: StatusUnhealthy,
				Error:  err.Error(),
			}
		} else {
			status.Checks["postgres"] = CheckResult{
				Status:    StatusHealthy,
				LatencyMs: time.Since(start).Milliseconds(),
			}
		}
	}

	// NATS check (only if configured)
	if c.natsURL != "" {
		start := time.Now()
		if err := c.checkNATS(checkCtx); err != nil {
			status.Status = StatusUnhealthy
			status.Checks["nats"] = CheckResult{
				Status: StatusUnhealthy,
				Error:  err.Error(),
			}
		} else {
			status.Checks["nats"] = CheckResult{
				Status:    StatusHealthy,
				LatencyMs: time.Since(start).Milliseconds(),
			}
		}
	}

	return status
}

// checkNATS creates a new NATS connection to verify connectivity.
func (c *Checker) checkNATS(ctx context.Context) error {
	conn, err := nats.Connect(c.natsURL, nats.Timeout(2*time.Second))
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// LiveHandler returns a Gin handler for liveness probes.
func (c *Checker) LiveHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// ReadyHandler returns a Gin handler for readiness probes.
func (c *Checker) ReadyHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := c.Check(ctx.Request.Context())

		httpStatus := http.StatusOK
		if status.Status != StatusHealthy {
			httpStatus = http.StatusServiceUnavailable
		}

		ctx.JSON(httpStatus, status)
	}
}

// IsHealthy returns true if all dependencies are healthy.
func (c *Checker) IsHealthy(ctx context.Context) bool {
	return c.Check(ctx).Status == StatusHealthy
}

// GRPCChecker implements grpchealth.Checker interface for gRPC health checking protocol.
type GRPCChecker struct {
	checker *Checker
}

// NewGRPCChecker creates a new gRPC health checker wrapping the given Checker.
func NewGRPCChecker(checker *Checker) *GRPCChecker {
	return &GRPCChecker{checker: checker}
}

// Check implements grpchealth.Checker interface.
func (g *GRPCChecker) Check(ctx context.Context, req *grpchealth.CheckRequest) (*grpchealth.CheckResponse, error) {
	if g.checker.IsHealthy(ctx) {
		return &grpchealth.CheckResponse{
			Status: grpchealth.StatusServing,
		}, nil
	}
	return &grpchealth.CheckResponse{
		Status: grpchealth.StatusNotServing,
	}, nil
}
