package middleware

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"strings"
)

// isOriginAllowed vérifie si une origine est autorisée avec credentials
func isOriginAllowed(origin string) bool {
	// Liste blanche d'origins autorisées avec credentials
	allowedOrigins := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://[::1]",  // IPv6 localhost
		"http://0.0.0.0",
		"http://host.docker.internal",
		"file://",  // Electron, Tauri, etc.
	}
	
	// Vérifier si l'origin est dans la liste blanche
	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(origin, allowed) {
			return true
		}
		// Sans le protocole
		if strings.HasPrefix(origin, strings.TrimPrefix(allowed, "http://")) {
			return true
		}
	}
	
	return false
}

// CORSMiddleware configure CORS avec support des credentials (cookies)
func CORSMiddleware() gin.HandlerFunc {
	// Cette ligne s'exécute PENDANT la registration du middleware  
	stderr := os.Stderr
	log.New(stderr, "[CORS-MIDDLEWARE-INIT] ", log.LstdFlags).Println("CORS middleware registered!")
	
	// Cette fonction s'exécute pour CHAQUE requête
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method
		path := c.Request.URL.Path
		
		log.Printf("[CORS-REQUEST] %s %s | Origin: %s\n", method, path, origin)
		
		// Vérifier si l'origin est autorisée
		allowCredentials := false
		if origin != "" && isOriginAllowed(origin) {
			// Origin reconnue et autorisée
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			allowCredentials = true
		} else if origin != "" && !isOriginAllowed(origin) {
			// Origin fourni mais pas autorisé - accepter sans credentials
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			allowCredentials = false
		} else {
			// Pas d'Origin header (requête directe, curl, etc.)
			// Par défaut : accepter sans credentials (plus sûr)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			allowCredentials = false
		}
		
		if allowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Max-Age", "3600")
		
		// Traiter les preflight requests
		if method == "OPTIONS" {
			log.Printf("[CORS-REQUEST] Responding to OPTIONS: %s\n", path)
			c.AbortWithStatus(204)
			return
		}
		
		// Continuer
		c.Next()
	}
}
