package controllers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthController struct {
	db        *pgxpool.Pool
	redis     *redis.Client
	startTime time.Time
	version   string
}

func NewHealthController(db *pgxpool.Pool, redis *redis.Client, startTime time.Time, version string) *HealthController {
	return &HealthController{
		db:        db,
		redis:     redis,
		startTime: startTime,
		version:   version,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	ServiceName string                 `json:"service_name"`
	Version     string                 `json:"version"`
	Uptime      string                 `json:"uptime"`
	GoVersion   string                 `json:"go_version"`
	NumRoutines int                    `json:"num_goroutine"`
	Checks      map[string]interface{} `json:"checks"`
	Memory      map[string]interface{} `json:"memory"`
}

// Health godoc
// @Summary Health check
// @Description Detailed health check with database, redis status, memory usage and uptime
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthController) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	overallStatus := "healthy"

	// Check PostgreSQL
	pgStatus := "ok"
	pgLatency := int64(0)
	start := time.Now()
	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			pgStatus = "error"
			overallStatus = "degraded"
		} else {
			pgLatency = time.Since(start).Milliseconds()
		}
	} else {
		pgStatus = "not_configured"
		overallStatus = "unhealthy"
	}

	// Check Redis
	redisStatus := "ok"
	redisLatency := int64(0)
	if h.redis != nil {
		start = time.Now()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			redisStatus = "error"
			// Redis error doesn't make service unhealthy (graceful degradation)
		} else {
			redisLatency = time.Since(start).Milliseconds()
		}
	} else {
		redisStatus = "disabled"
	}

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"status":        overallStatus,
		"timestamp":     time.Now().UTC(),
		"service_name":  "market-service",
		"version":       h.version,
		"uptime":        time.Since(h.startTime).String(),
		"go_version":    runtime.Version(),
		"num_goroutine": runtime.NumGoroutine(),
		"checks": gin.H{
			"postgres": gin.H{
				"status":     pgStatus,
				"latency_ms": pgLatency,
			},
			"redis": gin.H{
				"status":     redisStatus,
				"latency_ms": redisLatency,
			},
		},
		"memory": gin.H{
			"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(m.Sys) / 1024 / 1024,
			"num_gc":         m.NumGC,
		},
	})
}
