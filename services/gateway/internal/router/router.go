// Package router provides HTTP routing for the gateway.
package router

import (
	"fmt"
	"gateway-service/internal/config"
	"gateway-service/internal/middleware"
	"gateway-service/internal/proxy"
	"gateway-service/internal/registry"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configure toutes les routes de la gateway de manière dynamique.
// Chaque service enregistré dans la configuration crée automatiquement des routes
// qui sont proxifiées vers le service correspondant.
//
// Paramètres:
//   - r: L'engine Gin
//   - cfg: La configuration complète incluant la registry des services
//
// Routes créées automatiquement:
//   - GET/POST/PUT/DELETE /{service}/* -> Proxifiée vers le service cible
//   - Les routes avec auth_required=false ne nécessitent pas d'authentification
//   - Les routes avec auth_required=true nécessitent un JWT valide
//
// Routes spéciales (non proxifiées):
//   - GET /info -> Santé de la gateway et des services
//   - GET / -> Redirection vers la documentation
//   - GET /swagger/* -> Documentation Swagger
func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	fmt.Fprintf(os.Stderr, "[SETUP-ROUTES] Starting SetupRoutes function\n")
	
	// Ajouter le middleware CORS EN PREMIER (support cookies cross-origin)
	fmt.Fprintf(os.Stderr, "[SETUP-ROUTES] About to call CORSMiddleware()\n")
	r.Use(middleware.CORSMiddleware())
	fmt.Fprintf(os.Stderr, "[SETUP-ROUTES] CORSMiddleware() called successfully\n")

	// Appliquer le middleware de logging global
	r.Use(middleware.LoggerMiddleware())

	// Créer le proxy universel avec un client HTTP configurable
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	proxyHandler := proxy.NewUniversalProxyHandler(httpClient)

	// Route racine - redirection vers Swagger
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Route de santé - montre l'état de la gateway et des services
	r.GET("/info", createHealthCheckHandler(cfg.ServiceRegistry))

	// Créer les routes dynamiques pour chaque service enregistré
	createDynamicRoutes(r, cfg, proxyHandler)

	// Route catch-all pour les routes non trouvées
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})
}

// createDynamicRoutes crée les routes dynamiques pour tous les services enregistrés.
// Cette fonction itère sur la registry des services et crée pour chaque service
// une série de routes qui proxifient les requêtes HTTP vers le service cible.
//
// Paramètres:
//   - r: L'engine Gin
//   - cfg: La configuration avec la registry des services
//   - proxyHandler: Le proxy universel utilisé pour forwarder les requêtes
//
// Exemple:
// Pour un service "messages" avec prefix "/messages" et URL "http://message-service:3001":
//   - GET /messages/* -> Proxifiée vers http://message-service:3001/messages/*
//   - POST /messages/* -> Proxifiée vers http://message-service:3001/messages/*
//   - etc.
func createDynamicRoutes(r *gin.Engine, cfg *config.Config, proxyHandler *proxy.UniversalProxyHandler) {
	// Récupérer tous les services enregistrés
	services := cfg.ServiceRegistry.GetAll()

	// Créer des routes pour chaque service
	for _, service := range services {
		// Créer un groupe de routes avec le préfixe du service
		serviceGroup := r.Group(service.Prefix)

		// Si le service nécessite une authentification, ajouter le middleware JWT
		if service.AuthRequired {
			serviceGroup.Use(middleware.JWTMiddleware(cfg.JWT.Secret))
		}

		// Créer les routes catch-all pour tous les verbes HTTP
		// Le pattern "/*path" signifie que toutes les routes sous ce préfixe
		// seront proxifiées vers le service
		registrDynamicHTTPMethods(serviceGroup, service, proxyHandler)

		// Logger pour le debugging
		fmt.Printf("✓ Service %s enregistré sur %s -> %s (auth: %v)\n",
			service.Name, service.Prefix, service.URL, service.AuthRequired)
	}
}

// registrDynamicHTTPMethods crée les handlers pour tous les verbes HTTP (GET, POST, PUT, DELETE, etc.)
// pour un service donné. Cela permet de proxifier n'importe quelle requête vers le service cible.
//
// Paramètres:
//   - group: Le groupe de routes Gin pour le service
//   - service: Le service cible
//   - proxyHandler: Le proxy universel
//
// Verbes HTTP supportés:
// - GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
func registrDynamicHTTPMethods(group *gin.RouterGroup, service *registry.Service, proxyHandler *proxy.UniversalProxyHandler) {
	// Fonction générique pour créer un handler de proxification
	// Il construit l'URL cible en combinant l'URL du service avec le chemin restant
	createProxyHandler := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			// Récupérer le chemin restant après le préfixe du service
			path := c.Param("path")

			// Construire l'URL complète du service cible
			// Exemple: service.URL = "http://message-service:3001"
			//          service.Prefix = "/messages"
			//          path = "/auth/login" (le prefix a été retiré par Gin automatiquement)
			//          targetURL = "http://message-service:3001/messages/auth/login"
			targetURL := service.URL + service.Prefix + path

			// Proxifier la requête vers le service
			proxyHandler.ProxyRequest(c, targetURL)
		}
	}

	// Enregistrer les handlers pour tous les verbes HTTP courants
	// Chaque verbe crée une route "/*path" qui proxifie la requête
	httpMethods := map[string]func(string, ...gin.HandlerFunc) gin.IRoutes{
		"GET":     group.GET,
		"POST":    group.POST,
		"PUT":     group.PUT,
		"DELETE":  group.DELETE,
		"PATCH":   group.PATCH,
		"HEAD":    group.HEAD,
		"OPTIONS": group.OPTIONS,
	}

	for _, method := range []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"} {
		// Créer une copie de la fonction de handler pour éviter les closure issues
		handler := createProxyHandler()
		httpMethods[method]("/*path", handler)
	}
}

// createHealthCheckHandler crée un handler qui teste l'état de santé de la gateway
// et de tous les services enregistrés avec des requêtes HTTP réelles.
//
// Paramètres:
//   - registry: La registry des services
//
// Retour: Un handler Gin qui retourne l'information de santé au format JSON
//
// Pour chaque service, une requête GET est envoyée vers {service.URL}/info
// Le résultat retourne le status code de chaque service.
//
// Exemple de réponse:
//   {
//     "gateway": "ok",
//     "message": "200 OK",
//     "user": "200 OK",
//     "presence": "timeout"
//   }
func createHealthCheckHandler(reg *registry.ServiceRegistry) gin.HandlerFunc {
	return func(c *gin.Context) {
		services := reg.GetAll()
		results := make(map[string]string)
		results["gateway"] = "ok"

		// Client HTTP avec timeout court pour les health checks
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		// Tester chaque service avec une requête HTTP réelle
		for _, service := range services {
			healthURL := service.URL + "/info"
			resp, err := client.Get(healthURL)

			if err != nil {
				// Erreur de connexion, timeout, ou autre problème réseau
				if err.Error() == "context deadline exceeded" {
					results[service.Name] = "timeout"
				} else {
					results[service.Name] = fmt.Sprintf("error: %s", err.Error())
				}
				continue
			}

			// Fermer le body de la réponse
			_ = resp.Body.Close()

			// Retourner le status code
			statusText := "500 Error"
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				statusText = fmt.Sprintf("%d OK", resp.StatusCode)
			} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
				statusText = fmt.Sprintf("%d Redirect", resp.StatusCode)
			} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				statusText = fmt.Sprintf("%d Client Error", resp.StatusCode)
			} else {
				statusText = fmt.Sprintf("%d Server Error", resp.StatusCode)
			}

			results[service.Name] = statusText
		}

		c.JSON(http.StatusOK, results)
	}
}

