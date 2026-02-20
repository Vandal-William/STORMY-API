package service

import (
	"gateway-service/internal/config"
	"gateway-service/internal/domain"
	"gateway-service/pkg/client"
)

// HealthService handles health check operations
type HealthService struct {
	httpClient *client.HTTPClient
	config     *config.ServicesConfig
}

// NewHealthService creates a new health service
func NewHealthService(httpClient *client.HTTPClient, config *config.ServicesConfig) *HealthService {
	return &HealthService{
		httpClient: httpClient,
		config:     config,
	}
}

// GetGatewayInfo retrieves the health status of all services
func (s *HealthService) GetGatewayInfo() *domain.GatewayInfo {
	return &domain.GatewayInfo{
		Gateway:      "ok",
		Message:      s.httpClient.GetServiceStatus(s.config.MessageURL),
		Presence:     s.httpClient.GetServiceStatus(s.config.PresenceURL),
		User:         s.httpClient.GetServiceStatus(s.config.UserURL),
		Notification: s.httpClient.GetServiceStatus(s.config.NotificationURL),
		Moderation:   s.httpClient.GetServiceStatus(s.config.ModerationURL),
	}
}
