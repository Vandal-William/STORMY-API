package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// UniversalProxyHandler gère le proxy universel pour toutes les requêtes.
// Il copie intelligemment tous les headers, cookies, body et query parameters
// vers le service cible, puis retourne la réponse au client.
type UniversalProxyHandler struct {
	// httpClient est le client HTTP utilisé pour faire les requêtes
	httpClient *http.Client
}

// NewUniversalProxyHandler crée une nouvelle instance du proxy universel.
//
// Paramètres:
//   - httpClient: Le client HTTP à utiliser pour les requêtes vers les services
//
// Retour: Un pointeur sur UniversalProxyHandler
func NewUniversalProxyHandler(httpClient *http.Client) *UniversalProxyHandler {
	return &UniversalProxyHandler{
		httpClient: httpClient,
	}
}

// ProxyRequest proxifie une requête HTTP vers un service cible en copiant
// intelligemment tous les éléments de la requête originale:
// - Tous les headers HTTP
// - Tous les cookies
// - Le body (préservé tel quel)
// - Les query parameters
// - La méthode HTTP et le chemin
//
// Paramètres:
//   - c: Le contexte Gin contenant la requête originale
//   - targetURL: L'URL complète du service cible (ex: http://message-service:3001/messages/123)
//
// La fonction retourne automatiquement la réponse au client avec:
// - Le code de statut original
// - Tous les headers retournés par le service
// - Tous les cookies retournés par le service
// - Le body retourné par le service
func (h *UniversalProxyHandler) ProxyRequest(c *gin.Context, targetURL string) {
	// Valider l'URL cible
	if _, err := url.Parse(targetURL); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid target URL"})
		return
	}

	// Créer la requête vers le service cible
	req, err := h.buildProxyRequest(c, targetURL)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to build proxy request"})
		return
	}

	// Envoyer la requête vers le service cible
	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "service unavailable"})
		return
	}
	defer resp.Body.Close()

	// Copier les headers de la réponse
	h.copyResponseHeaders(c, resp)

	// Copier les cookies de la réponse
	h.copyResponseCookies(c, resp)

	// Lire et renvoyer le body
	body, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// buildProxyRequest construit une requête HTTP vers le service cible en copiant
// tous les éléments pertinents de la requête originale.
//
// Paramètres:
//   - c: Le contexte Gin contenant la requête originale
//   - targetURL: L'URL du service cible
//
// Retour: Une requête HTTP prête à être exécutée et une erreur le cas échéant
//
// Cette fonction gère:
// - La lecture et la copie du body
// - La copie de tous les headers HTTP
// - La copie des cookies
// - Les query parameters
func (h *UniversalProxyHandler) buildProxyRequest(c *gin.Context, targetURL string) (*http.Request, error) {
	// Lire le body de la requête originale
	var bodyReader io.Reader
	if c.Request.Body != nil {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			return nil, err
		}
		// Remettre le body original en cas de besoin ultérieur
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		bodyReader = bytes.NewBuffer(body)
	}

	// Créer la nouvelle requête avec la méthode et l'URL cible
	req, err := http.NewRequest(c.Request.Method, targetURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Copier TOUS les headers de la requête originale
	h.copyRequestHeaders(c.Request, req)

	// Copier TOUS les cookies de la requête originale
	h.copyRequestCookies(c.Request, req)

	// Copier les query parameters
	req.URL.RawQuery = c.Request.URL.RawQuery

	// S'assurer explicitement que le JWT Authorization est transmis au service cible
	if authHeader := c.Request.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	return req, nil
}

// copyRequestHeaders copie tous les headers HTTP de la requête originale
// vers la requête du proxy, en excluant les headers "hop-by-hop" qui ne
// doivent pas être transmis (Host, Connection, Transfer-Encoding, etc.).
//
// Paramètres:
//   - src: La requête source
//   - dst: La requête destination (proxy)
func (h *UniversalProxyHandler) copyRequestHeaders(src *http.Request, dst *http.Request) {
	// Headers "hop-by-hop" qui ne doivent pas être copiés
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"TE":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
	}

	// Copier tous les headers sauf ceux "hop-by-hop"
	for key, values := range src.Header {
		if !hopByHopHeaders[key] {
			for _, value := range values {
				dst.Header.Add(key, value)
			}
		}
	}
}

// copyRequestCookies copie tous les cookies de la requête originale
// vers la requête du proxy en utilisant le header Cookie standard.
//
// Paramètres:
//   - src: La requête source
//   - dst: La requête destination (proxy)
func (h *UniversalProxyHandler) copyRequestCookies(src *http.Request, dst *http.Request) {
	// Les cookies sont déjà dans src.Cookies(), on les copie tous
	for _, cookie := range src.Cookies() {
		dst.AddCookie(cookie)
	}
}

// copyResponseHeaders copie tous les headers de la réponse du service cible
// vers la réponse finale au client, en excluant les headers "hop-by-hop".
//
// Paramètres:
//   - c: Le contexte Gin
//   - resp: La réponse du service cible
func (h *UniversalProxyHandler) copyResponseHeaders(c *gin.Context, resp *http.Response) {
	// Headers "hop-by-hop" qui ne doivent pas être copiés
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":          true,
		"Proxy-Authenticate":  true,
		"Proxy-Authorization": true,
		"TE":                  true,
		"Trailers":            true,
		"Transfer-Encoding":   true,
		"Upgrade":             true,
		"Content-Length":      true, // Gin gère la longueur du contenu
	}

	// Copier tous les headers sauf ceux "hop-by-hop"
	for key, values := range resp.Header {
		if !hopByHopHeaders[key] {
			for _, value := range values {
				c.Header(key, value)
			}
		}
	}
}

// copyResponseCookies copie tous les cookies de la réponse du service cible
// vers la réponse finale au client.
//
// Paramètres:
//   - c: Le contexte Gin
//   - resp: La réponse du service cible
func (h *UniversalProxyHandler) copyResponseCookies(c *gin.Context, resp *http.Response) {
	// Copier tous les cookies de la réponse du service
	for _, cookie := range resp.Cookies() {
		c.SetCookie(
			cookie.Name,
			cookie.Value,
			cookie.MaxAge,
			cookie.Path,
			cookie.Domain,
			cookie.Secure,
			cookie.HttpOnly,
		)
	}
}
