package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SwaggerHandler handles Swagger UI requests
type SwaggerHandler struct{}

// NewSwaggerHandler creates a new swagger handler
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// GetSwaggerUI returns the Swagger UI HTML page
func (h *SwaggerHandler) GetSwaggerUI(c *gin.Context) {
	// If the path is just "/", redirect to "/swagger"
	if c.Request.URL.Path == "/" {
		c.Redirect(http.StatusMovedPermanently, "/swagger")
		return
	}

	// Serve Swagger UI
	html := `
<!DOCTYPE html>
<html>
  <head>
    <title>Gateway Service API</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" rel="stylesheet">
    <style>
      body { margin: 0; padding: 0; }
      .topbar { background-color: #fafafa; padding: 10px 0; border-bottom: 1px solid #e0e0e0; }
      .swagger-ui { max-width: 1200px; margin: 0 auto; }
    </style>
  </head>
  <body>
    <div class="topbar">
      <div class="container">
        <h1 style="margin: 10px 0; font-size: 24px;">📡 Gateway Service API</h1>
        <p style="margin: 0; color: #666;">JWT Authentication with Cookie Management</p>
      </div>
    </div>
    <div class="container swagger-ui" id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-standalone-preset.js"></script>
    <script>
      const ui = SwaggerUIBundle({
        url: "/api/swagger.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout",
        defaultModelsExpandDepth: 1,
        defaultModelExpandDepth: 1
      })
      window.onload = function() {
        window.ui = ui
      }
    </script>
  </body>
</html>
`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// GetSwaggerJSON returns the Swagger OpenAPI specification
func (h *SwaggerHandler) GetSwaggerJSON(c *gin.Context) {
	swaggerJSON := `{
  "swagger": "2.0",
  "info": {
    "title": "Gateway Service API",
    "description": "Gateway service that proxies requests to downstream services with JWT authentication",
    "version": "1.0"
  },
  "host": "localhost:8080",
  "basePath": "/",
  "schemes": ["http", "https"],
  "consumes": ["application/json"],
  "produces": ["application/json"],
  "securityDefinitions": {
    "CookieAuth": {
      "type": "apiKey",
      "name": "access_token",
      "in": "cookie"
    }
  },
  "paths": {
    "/info": {
      "get": {
        "summary": "Health check",
        "description": "Returns the health status of the gateway and all downstream services",
        "responses": {
          "200": {
            "description": "Service is healthy"
          }
        }
      }
    },
    "/auth/register": {
      "post": {
        "summary": "Register a new user",
        "description": "Creates a new user account and returns JWT token in cookie",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "properties": {
                "username": { "type": "string", "minLength": 3 },
                "password": { "type": "string", "minLength": 8 },
                "phone": { "type": "string", "minLength": 6 },
                "email": { "type": "string", "format": "email" }
              }
            }
          }
        ],
        "responses": {
          "201": { "description": "User registered successfully" },
          "400": { "description": "Bad request" },
          "409": { "description": "Username or phone already taken" }
        }
      }
    },
    "/auth/login": {
      "post": {
        "summary": "Login user",
        "description": "Authenticates user and returns JWT token in cookie",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "object",
              "properties": {
                "username": { "type": "string" },
                "password": { "type": "string" }
              }
            }
          }
        ],
        "responses": {
          "200": { "description": "Login successful" },
          "401": { "description": "Invalid credentials" }
        }
      }
    },
    "/conversations": {
      "post": {
        "summary": "Create a new conversation",
        "security": [{"CookieAuth": []}],
        "responses": {
          "201": { "description": "Conversation created" },
          "401": { "description": "Unauthorized" }
        }
      }
    },
    "/conversations/{id}": {
      "get": {
        "summary": "Get conversation details",
        "security": [{"CookieAuth": []}],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "type": "string" }
        ],
        "responses": {
          "200": { "description": "Conversation details" },
          "401": { "description": "Unauthorized" }
        }
      }
    },
    "/messages": {
      "post": {
        "summary": "Create a new message",
        "security": [{"CookieAuth": []}],
        "responses": {
          "201": { "description": "Message created" },
          "401": { "description": "Unauthorized" }
        }
      }
    }
  }
}`
	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, swaggerJSON)
}
