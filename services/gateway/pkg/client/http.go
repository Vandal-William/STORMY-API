package client

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// HTTPClient wraps http.Client for making requests
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client with default timeouts
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Do sends an HTTP request with the specified method, URL, content-type and body
func (c *HTTPClient) Do(method, url, contentType string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.client.Do(req)
}

// GetServiceStatus checks the health of a service
func (c *HTTPClient) GetServiceStatus(baseURL string) string {
	if baseURL == "" {
		return "not-configured"
	}

	resp, err := c.client.Get(baseURL + "/info")
	if err != nil || resp == nil {
		return "unreachable"
	}
	defer resp.Body.Close()

	return resp.Status
}
