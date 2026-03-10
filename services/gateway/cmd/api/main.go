package main

import (
	"gateway-service/internal/config"
	"gateway-service/internal/router"
	"log"

	"github.com/gin-gonic/gin"
)

// main est le point d'entrée de la gateway.
//
// Elle charge la configuration depuis les variables d'environnement et le fichier YAML,
// initialize l'engine Gin, configure toutes les routes de manière dynamique,
// et démarre le serveur HTTP.
//
// Configuration:
//   - Variables d'environnement: PORT, HOST, JWT_SECRET
//   - Fichier YAML: config/services.yaml (liste des services)
//
// Dépendances:
//   - File: config/services.yaml
//   - Environment variables: PORT, HOST, JWT_SECRET
//
// Logs:
//   - Affiche l'adresse du serveur au démarrage
//   - Affiche les services enregistrés
//   - Affiche les erreurs en cas d'échec
func main() {
	// Charger la configuration depuis l'environnement et le YAML
	cfg := config.Load()

	// Afficher les informations de démarrage
	log.Printf("========================================\n")
	log.Printf("Gateway API - Démarrage\n")
	log.Printf("========================================\n")
	log.Printf("Serveur: %s\n", cfg.Server.GetAddr())
	log.Printf("Services enregistrés: %d\n", len(cfg.Services))

	// Afficher la liste des services enregistrés
	for i, service := range cfg.Services {
		authStr := "❌"
		if service.AuthRequired {
			authStr = "✓"
		}
		log.Printf("  %d. %s - %s [Auth: %s]\n", i+1, service.Name, service.URL, authStr)
	}
	log.Printf("========================================\n")

	// Créer l'engine Gin
	// Utilisez gin.DebugMode ou gin.ReleaseMode selon vos besoins
	r := gin.Default()

	// Configurer toutes les routes de manière dynamique
	router.SetupRoutes(r, cfg)

	// Démarrer le serveur HTTP
	// L'adresse est définie dans la configuration
	log.Printf("Démarrage du serveur sur %s\n", cfg.Server.GetAddr())
	if err := r.Run(cfg.Server.GetAddr()); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v\n", err)
	}
}
