package handler

import (
	"gateway-service/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health-related requests
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{
		healthService: healthService,
	}
}

// GetInfo returns the gateway and downstream services health info
func (h *HealthHandler) GetInfo(c *gin.Context) {
	info := h.healthService.GetGatewayInfo()
	c.JSON(http.StatusOK, info)
}
