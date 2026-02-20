package unit

import (
	"gateway-service/internal/config"
	"gateway-service/internal/domain"
	"gateway-service/internal/service"
	"testing"
	"time"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	responses map[string]string
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]string),
	}
}

func (m *MockHTTPClient) GetServiceStatus(baseURL string) string {
	if response, ok := m.responses[baseURL]; ok {
		return response
	}
	return domain.StatusUnreachable
}

func TestGetGatewayInfo(t *testing.T) {
	// Setup
	mockClient := NewMockHTTPClient()
	mockClient.responses["http://user:3000"] = "200 OK"
	mockClient.responses["http://message:3001"] = "200 OK"
	mockClient.responses[""] = domain.StatusNotConfigured

	cfg := &config.ServicesConfig{
		UserURL:         "http://user:3000",
		MessageURL:      "http://message:3001",
		PresenceURL:     "",
		NotificationURL: "",
		ModerationURL:   "",
	}

	healthService := service.NewHealthService(mockClient, cfg)

	// Execute
	info := healthService.GetGatewayInfo()

	// Verify
	if info.Gateway != "ok" {
		t.Errorf("expected gateway status 'ok', got '%s'", info.Gateway)
	}

	if info.User != "200 OK" {
		t.Errorf("expected user status '200 OK', got '%s'", info.User)
	}

	if info.Presence != domain.StatusNotConfigured {
		t.Errorf("expected presence status 'not-configured', got '%s'", info.Presence)
	}
}

func TestHealthServiceCreation(t *testing.T) {
	mockClient := NewMockHTTPClient()
	cfg := &config.ServicesConfig{}

	service := service.NewHealthService(mockClient, cfg)

	if service == nil {
		t.Error("expected service to be created, got nil")
	}
}
