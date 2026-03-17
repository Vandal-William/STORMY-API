// Package registry provides service registration for the gateway.
package registry

import (
	"fmt"
	"sync"
)

// Service représente un service en aval enregistré dans la gateway.
// Chaque service a une URL de base et un préfixe de route.
type Service struct {
	// Name est le nom unique du service (ex: "messages", "users")
	Name string

	// URL est l'adresse complète du service (ex: "http://message-service:3001")
	URL string

	// Prefix est le préfixe de route pour accéder à ce service (ex: "/messages")
	Prefix string

	// Description explique le rôle du service
	Description string

	// AuthRequired indique si l'authentification JWT est requise pour ce service
	AuthRequired bool
}

// ServiceRegistry gère l'enregistrement et la découverte des services.
// Il permet à la gateway de maintenir une liste à jour de tous les services
// disponibles et de router les requêtes vers le bon service.
type ServiceRegistry struct {
	// services est une map qui associe les préfixes de route aux services
	// Exemple: "/messages" -> Service{Name: "messages", ...}
	services map[string]*Service

	// servicesByName est une map qui associe les noms de service aux services
	// pour un accès rapide par nom
	servicesByName map[string]*Service

	// mutex protège l'accès concurrent à la registry
	mutex sync.RWMutex
}

// NewServiceRegistry crée une nouvelle instance vide de la registry des services.
// La registry est thread-safe et peut être utilisée de manière concurrente.
//
// Retour: Un pointeur sur ServiceRegistry initialisé et vide
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services:       make(map[string]*Service),
		servicesByName: make(map[string]*Service),
	}
}

// Register enregistre un nouveau service dans la registry.
// Si un service avec le même préfixe ou le même nom existe déjà,
// il sera remplacé par le nouveau.
//
// Paramètres:
//   - service: Le service à enregistrer
//
// Retour: Une erreur si le service est invalide (URL ou préfixe vide)
//
// Exemple:
//   registry.Register(&Service{
//       Name: "messages",
//       URL: "http://message-service:3001",
//       Prefix: "/messages",
//       AuthRequired: true,
//   })
func (sr *ServiceRegistry) Register(service *Service) error {
	// Valider le service
	if err := service.Validate(); err != nil {
		return err
	}

	// Acquérir le verrou d'écriture pour accès exclusif
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	// Enregistrer le service tant par préfixe que par nom
	sr.services[service.Prefix] = service
	sr.servicesByName[service.Name] = service

	return nil
}

// FindByPrefix cherche un service par son préfixe de route.
// Cette fonction est thread-safe et utilise des verrous en lecture.
//
// Paramètres:
//   - prefix: Le préfixe de route (ex: "/messages")
//
// Retour:
//   - Le service correspondant s'il existe
//   - true si le service a été trouvé, false sinon
//
// Exemple:
//   if service, found := registry.FindByPrefix("/messages"); found {
//       fmt.Println("Service trouvé:", service.URL)
//   }
func (sr *ServiceRegistry) FindByPrefix(prefix string) (*Service, bool) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	service, exists := sr.services[prefix]
	return service, exists
}

// FindByName cherche un service par son nom.
// Cette fonction est thread-safe et utilise des verrous en lecture.
//
// Paramètres:
//   - name: Le nom du service (ex: "messages")
//
// Retour:
//   - Le service correspondant s'il existe
//   - true si le service a été trouvé, false sinon
//
// Exemple:
//   if service, found := registry.FindByName("messages"); found {
//       fmt.Println("Service trouvé:", service.URL)
//   }
func (sr *ServiceRegistry) FindByName(name string) (*Service, bool) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	service, exists := sr.servicesByName[name]
	return service, exists
}

// GetAll retourne une copie de tous les services enregistrés.
// Utile pour lister tous les services, notamment pour le health check.
//
// Retour: Une slice de tous les services enregistrés
//
// Exemple:
//   services := registry.GetAll()
//   for _, svc := range services {
//       fmt.Println(svc.Name)
//   }
func (sr *ServiceRegistry) GetAll() []*Service {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()

	services := make([]*Service, 0, len(sr.services))
	for _, service := range sr.services {
		services = append(services, service)
	}
	return services
}

// Unregister supprime un service de la registry par son préfixe.
// Cette fonction est thread-safe.
//
// Paramètres:
//   - prefix: Le préfixe du service à supprimer
//
// Retour: true si un service a été supprimé, false si aucun service ne correspondait
func (sr *ServiceRegistry) Unregister(prefix string) bool {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	service, exists := sr.services[prefix]
	if !exists {
		return false
	}

	// Supprimer à la fois par préfixe et par nom
	delete(sr.services, prefix)
	delete(sr.servicesByName, service.Name)
	return true
}

// Validate vérifie que le service a tous les champs requis.
// Retourne une erreur si une validation échoue.
//
// Retour: Une erreur descriptive si la validation échoue, nil sinon
func (s *Service) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if s.URL == "" {
		return fmt.Errorf("service URL is required")
	}
	if s.Prefix == "" {
		return fmt.Errorf("service prefix is required")
	}
	return nil
}

// String retourne une représentation textuelle du service pour le logging.
//
// Retour: Une chaîne de caractères décrivant le service
func (s *Service) String() string {
	return fmt.Sprintf("Service{Name: %s, Prefix: %s, URL: %s, AuthRequired: %v}",
		s.Name, s.Prefix, s.URL, s.AuthRequired)
}
