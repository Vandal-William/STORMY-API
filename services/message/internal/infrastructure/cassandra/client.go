package cassandra

import (
	"fmt"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

// Client wrapper pour gocql avec gestion des erreurs et reconnexion
type Client struct {
	session *gocql.Session
	mu      sync.RWMutex
	closed  bool
}

// Config pour la connexion Cassandra
type Config struct {
	Hosts            []string
	Port             int
	Keyspace         string
	Username         string
	Password         string
	ConsistencyLevel string
	Timeout          time.Duration
	ConnectTimeout   time.Duration
}

// NewClient crée une nouvelle connexion Cassandra
func NewClient(cfg Config) (*Client, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Port = cfg.Port
	cluster.Keyspace = cfg.Keyspace
	cluster.Timeout = cfg.Timeout
	cluster.ConnectTimeout = cfg.ConnectTimeout

	// Configuration de l'authentification si nécessaire
	if cfg.Username != "" && cfg.Password != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	// Configuration de la cohérence
	switch cfg.ConsistencyLevel {
	case "ONE":
		cluster.Consistency = gocql.One
	case "TWO":
		cluster.Consistency = gocql.Two
	case "THREE":
		cluster.Consistency = gocql.Three
	case "QUORUM":
		cluster.Consistency = gocql.Quorum
	case "LOCAL_QUORUM":
		cluster.Consistency = gocql.LocalQuorum
	default:
		cluster.Consistency = gocql.LocalOne
	}

	// Politique de reconnexion
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{
		NumRetries: 3,
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create cassandra session: %w", err)
	}

	return &Client{
		session: session,
		closed:  false,
	}, nil
}

// GetSession retourne la session active
func (c *Client) GetSession() *gocql.Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.session
}

// Close ferme la connexion Cassandra
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.session.Close()
	c.closed = true
	return nil
}

// HealthCheck vérifie que la connexion est établie
func (c *Client) HealthCheck() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("cassandra client is closed")
	}

	query := c.session.Query("SELECT now() FROM system.local LIMIT 1")
	err := query.Exec()
	if err != nil {
		return fmt.Errorf("cassandra health check failed: %w", err)
	}
	return nil
}

// ExecuteQuery exécute une requête personnalisée
func (c *Client) ExecuteQuery(stmt string, values ...interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("cassandra client is closed")
	}

	return c.session.Query(stmt, values...).Exec()
}

// GetValue récupère une seule valeur
func (c *Client) GetValue(stmt string, dest interface{}, values ...interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("cassandra client is closed")
	}

	return c.session.Query(stmt, values...).Scan(dest)
}

// GetValues récupère plusieurs valeurs
func (c *Client) GetValues(stmt string, values ...interface{}) (*gocql.Iter, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("cassandra client is closed")
	}

	return c.session.Query(stmt, values...).Iter(), nil
}
