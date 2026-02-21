package domain

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

// GatewayInfo represents the gateway health info
type GatewayInfo struct {
	Gateway      string `json:"gateway"`
	Message      string `json:"message"`
	Presence     string `json:"presence"`
	User         string `json:"user"`
	Notification string `json:"notification"`
	Moderation   string `json:"moderation"`
}

// ServiceStatus constants
const (
	StatusOk           = "200 OK"
	StatusUnreachable  = "unreachable"
	StatusNotConfigured = "not-configured"
)
