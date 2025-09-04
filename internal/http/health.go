package http

import "github.com/gin-gonic/gin"

// HealthHandler provides liveness/readiness endpoints.
type HealthHandler struct{ dbPing func() error }

func NewHealthHandler(dbPing func() error) *HealthHandler { return &HealthHandler{dbPing: dbPing} }

func (h *HealthHandler) Register(r *gin.Engine) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/readyz", func(c *gin.Context) {
		if h.dbPing != nil && h.dbPing() != nil {
			c.JSON(503, gin.H{"status": "degraded"})
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
}
