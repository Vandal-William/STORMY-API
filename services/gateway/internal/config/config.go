package config

import (
	"fmt"
	"gateway-service/internal/registry"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config contient toute la configuration pour la gateway.
// Elle regroupe la configuration du serveur, des services en aval,
// et les paramètres de sécurité comme le JWT.
type Config struct {
	// Server est la configuration du serveur HTTP
	Server ServerConfig

	// Services est la liste des services en aval découverte depuis la YAML
	Services []*registry.Service

	// JWT contient la clé secrète pour valider les tokens JWT
	JWT JWTConfig

	// ServiceRegistry est le registre des services utilisé pour les lookups
	ServiceRegistry *registry.ServiceRegistry
}

// ServerConfig contient la configuration du serveur HTTP.
type ServerConfig struct {
	// Port est le port d'écoute du serveur (par défaut: 8080)
	Port string

	// Host est l'adresse IP d'écoute (par défaut: 0.0.0.0 pour toutes les interfaces)
	Host string

	// ReadTimeout est le délai maximum pour lire une requête
	ReadTimeout time.Duration

	// WriteTimeout est le délai maximum pour écrire une réponse
	WriteTimeout time.Duration

	// IdleTimeout est le délai avant fermeture d'une connexion inactive
	IdleTimeout time.Duration
}

// JWTConfig contient la configuration JWT.
type JWTConfig struct {
	// Secret est la clé secrète utilisée pour valider les tokens JWT
	Secret string
}

// ServicesYAML est la structure pour désérialiser le fichier services.yaml
type ServicesYAML struct {
	// Services est la liste des services définis dans le YAML
	Services []*registry.Service `yaml:"services"`
}

// Load charge la configuration depuis les variables d'environnement et le fichier YAML.
// Elle retourne la configuration complètement initialisée avec les services enregistrés.
//
// La fonction lit:
// - Les variables d'environnement pour les paramètres du serveur et JWT
// - Le fichier config/services.yaml pour la liste des services
//
// Retour: Un pointeur sur Config avec tous les paramètres initialisés et validés
//
// Exemple:
//   cfg := config.Load()
//   fmt.Println("Gateway lancée sur", cfg.Server.GetAddr())
func Load() *Config {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		JWT: JWTConfig{
			Secret: os.Getenv("JWT_SECRET"),
		},
		ServiceRegistry: registry.NewServiceRegistry(),
	}

	// Charger les services depuis le fichier YAML
	services, err := loadServicesFromYAML("config/services.yaml")
	if err != nil {
		// Si le fichier YAML n'existe pas, utiliser une configuration par défaut
		// avec les variables d'environnement
		loadServicesFromEnv(cfg)
	} else {
		// Enregistrer tous les services de la YAML dans la registry
		cfg.Services = services
		for _, service := range services {
			if err := cfg.ServiceRegistry.Register(service); err != nil {
				fmt.Printf("Erreur lors de l'enregistrement du service %s: %v\n", service.Name, err)
			}
		}
	}

	return cfg
}

// loadServicesFromYAML charge l'ensemble des services depuis un fichier YAML.
// Le fichier doit contenir une liste de services avec leurs paramètres.
//
// Paramètres:
//   - filepath: Le chemin vers le fichier YAML (ex: "config/services.yaml")
//
// Retour:
//   - Une slice de Services chargés depuis le YAML
//   - Une erreur si le fichier n'existe pas ou ne peut pas être parsé
//
// Format attendu:
//   services:
//     - name: messages
//       url: http://message-service:3001
//       prefix: /messages
//       description: Service de gestion des messages
//       auth_required: true
func loadServicesFromYAML(filepath string) ([]*registry.Service, error) {
	// Lire le fichier YAML
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("impossible de lire le fichier %s: %w", filepath, err)
	}

	// Parser le YAML
	var servicesYAML ServicesYAML
	if err := yaml.Unmarshal(data, &servicesYAML); err != nil {
		return nil, fmt.Errorf("erreur lors du parsing du YAML: %w", err)
	}

	return servicesYAML.Services, nil
}

// loadServicesFromEnv charge les services directement depuis les variables
// d'environnement. C'est une solution de secours si le fichier YAML ne
// peut pas être chargé.
//
// Les variables d'environnement attendues sont:
// - USER_SERVICE_URL
// - MESSAGE_SERVICE_URL
// - PRESENCE_SERVICE_URL
// - NOTIFICATION_SERVICE_URL
// - MODERATION_SERVICE_URL
//
// Paramètres:
//   - cfg: La configuration à remplir
func loadServicesFromEnv(cfg *Config) {
	// Services définis via les variables d'environnement (héritage de l'ancienne config)
	defaultServices := []*registry.Service{
		{
			Name:         "users",
			URL:          os.Getenv("USER_SERVICE_URL"),
			Prefix:       "/users",
			Description:  "Service de gestion des utilisateurs",
			AuthRequired: false,
		},
		{
			Name:         "messages",
			URL:          os.Getenv("MESSAGE_SERVICE_URL"),
			Prefix:       "/messages",
			Description:  "Service de gestion des messages",
			AuthRequired: true,
		},
		{
			Name:         "presence",
			URL:          os.Getenv("PRESENCE_SERVICE_URL"),
			Prefix:       "/presence",
			Description:  "Service de détection de présence",
			AuthRequired: true,
		},
		{
			Name:         "notification",
			URL:          os.Getenv("NOTIFICATION_SERVICE_URL"),
			Prefix:       "/notification",
			Description:  "Service de notifications",
			AuthRequired: true,
		},
		{
			Name:         "moderation",
			URL:          os.Getenv("MODERATION_SERVICE_URL"),
			Prefix:       "/moderation",
			Description:  "Service de modération",
			AuthRequired: true,
		},
		{
			Name:         "auth",
			URL:          os.Getenv("USER_SERVICE_URL"),
			Prefix:       "/auth",
			Description:  "Service d'authentification",
			AuthRequired: false,
		},
	}

	// Enregistrer uniquement les services ayant une URL configurée
	for _, service := range defaultServices {
		if service.URL != "" {
			cfg.Services = append(cfg.Services, service)
			if err := cfg.ServiceRegistry.Register(service); err != nil { return nil, err }
		}
	}
}

// getEnv récupère une variable d'environnement ou retourne une valeur par défaut.
// Utile pour les valeurs optionnelles avec des défauts sensibles.
//
// Paramètres:
//   - key: Le nom de la variable d'environnement
//   - defaultVal: La valeur par défaut si la variable n'existe pas
//
// Retour: La valeur de la variable ou la valeur par défaut
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

// GetAddr retourne l'adresse complète du serveur au format "host:port".
// Utilisé pour démarrer le serveur HTTP avec gin.
//
// Retour: Une chaîne de caractères au format "host:port" (ex: "0.0.0.0:8080")
//
// Exemple:
//   addr := cfg.Server.GetAddr()
//   r.Run(addr)
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}
