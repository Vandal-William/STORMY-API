package middleware

import (
    "errors"
    "fmt"
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// JWTClaims contient les claims Customisé du JWT
type JWTClaims struct {
    UserID   string `json:"user_id"`    // UUID string from gateway
    Username string `json:"username"`
    Email    string `json:"email"`
    jwt.RegisteredClaims
}

// JWTMiddleware crée un middleware pour la validation JWT
// Supporte les deux mécanismes : Authorization header ET cookie
func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        var tokenString string
        var err error

        // [DEBUG] Log tous les headers et cookies reçus
        fmt.Printf("\n=== MESSAGE-SERVICE JWT MIDDLEWARE ===\n")
        fmt.Printf("[MSG] Request: %s %s\n", c.Request.Method, c.Request.RequestURI)
        fmt.Printf("[MSG] Headers reçus du gateway:\n")
        for name, values := range c.Request.Header {
            if name == "Authorization" {
                fmt.Printf("  - %s: %s...\n", name, values[0][:50])
            } else if name == "Cookie" {
                fmt.Printf("  - %s: %v\n", name, values)
            } else {
                fmt.Printf("  - %s: %v\n", name, values)
            }
        }
        fmt.Printf("[MSG] Tous les cookies:\n")
        for _, cookie := range c.Request.Cookies() {
            if len(cookie.Value) > 30 {
                fmt.Printf("  - %s: %s...\n", cookie.Name, cookie.Value[:30])
            } else {
                fmt.Printf("  - %s: %s\n", cookie.Name, cookie.Value)
            }
        }
        fmt.Printf("========================================\n\n")

        // 1. Essayer d'abord le cookie "access_token" (du gateway)
        tokenString, err = c.Cookie("access_token")
        if err != nil {
            // 2. Essayer cookie "authorization"
            tokenString, err = c.Cookie("authorization")
            if err != nil {
                // 3. Si pas de cookie, essayer le header Authorization
                authHeader := c.GetHeader("Authorization")
                if authHeader == "" {
                    c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization cookie or header"})
                    c.Abort()
                    return
                }

                // Extraire le token du format "Bearer <token>"
                parts := strings.Split(authHeader, " ")
                if len(parts) != 2 || parts[0] != "Bearer" {
                    c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
                    c.Abort()
                    return
                }
                tokenString = parts[1]
            }
        }

        // Valider et parser le token
        claims := &JWTClaims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            // Vérifier la méthode de signature
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("unexpected signing method")
            }
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            c.Abort()
            return
        }

        // Stocker l'userID dans le contexte
        userID := claims.UserID
        if userID == "" && claims.Subject != "" {
            // Fallback sur RegisteredClaims.Subject ("sub") si "user_id" n'existe pas
            userID = claims.Subject
        }
        
        c.Set("user_id", userID)
        c.Set("username", claims.Username)
        c.Set("email", claims.Email)
        c.Set("claims", claims)

        c.Next()
    }
}

// GetUserIDFromContext extrait l'userID du contexte
func GetUserIDFromContext(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", errors.New("user_id is not string")
	}

	return userIDStr, nil
}
// GetUsernameFromContext extrait l'username du contexte
func GetUsernameFromContext(c *gin.Context) (string, error) {
    username, exists := c.Get("username")
    if !exists {
        return "", errors.New("username not found in context")
    }

    usernameStr, ok := username.(string)
    if !ok {
        return "", errors.New("username is not string")
    }

    return usernameStr, nil
}

// GetClaimsFromContext retourne les claims JWT complets
func GetClaimsFromContext(c *gin.Context) (*JWTClaims, error) {
    claims, exists := c.Get("claims")
    if !exists {
        return nil, errors.New("claims not found in context")
    }

    jwtClaims, ok := claims.(*JWTClaims)
    if !ok {
        return nil, errors.New("claims is not of type JWTClaims")
    }

    // S'assurer que UserID est rempli (fallback sur Subject si vide)
    if jwtClaims.UserID == "" && jwtClaims.Subject != "" {
        jwtClaims.UserID = jwtClaims.Subject
    }

    return jwtClaims, nil
}