package handler

import (
	"gateway-service/pkg/client"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication service proxy requests
type AuthHandler struct {
	httpClient *client.HTTPClient
	userURL    string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(httpClient *client.HTTPClient, userURL string) *AuthHandler {
	return &AuthHandler{
		httpClient: httpClient,
		userURL:    userURL,
	}
}

// Register proxies POST /auth/register request to user service
// @Summary      Register a new user
// @Description  Creates a new user account and returns JWT token in cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "Registration data"
// @Success      201  {object} map[string]string
// @Failure      400  {object} ErrorResponse
// @Failure      409  {object} ErrorResponse "Username or phone already taken"
// @Failure      500  {object} ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Forward request to user service
	resp, err := h.httpClient.Do("POST", h.userURL+"/auth/register",
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response body"})
		return
	}

	// Copy Set-Cookie headers from user service response
	for _, cookie := range resp.Cookies() {
		c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	}

	// Return response with same status code
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// Login proxies POST /auth/login request to user service
// @Summary      Login user
// @Description  Authenticates user and returns JWT token in cookie
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "Login credentials"
// @Success      200  {object} map[string]string
// @Failure      400  {object} ErrorResponse
// @Failure      401  {object} ErrorResponse "Invalid credentials"
// @Failure      500  {object} ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Forward request to user service
	resp, err := h.httpClient.Do("POST", h.userURL+"/auth/login",
		c.Request.Header.Get("Content-Type"), body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response body"})
		return
	}

	// Copy Set-Cookie headers from user service response
	for _, cookie := range resp.Cookies() {
		c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	}

	// Return response with same status code
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
