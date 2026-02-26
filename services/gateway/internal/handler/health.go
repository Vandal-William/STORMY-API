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
// @Summary      Health Check
// @Description  Returns the health status of the gateway and all downstream services
// @Tags         health
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Router       /info [get]
func (h *HealthHandler) GetInfo(c *gin.Context) {
	info := h.healthService.GetGatewayInfo()
	c.JSON(http.StatusOK, info)
}
