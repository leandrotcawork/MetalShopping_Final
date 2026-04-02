package erp_runtime

import (
	"fmt"
	"sync"
)

// Registry holds registered ERP connectors keyed by their type identifier.
type Registry struct {
	mu         sync.RWMutex
	connectors map[string]Connector
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{connectors: make(map[string]Connector)}
}

// Register adds a connector to the registry.
func (r *Registry) Register(c Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors[c.Type()] = c
}

// Get returns the connector for the given type, or an error if not registered.
func (r *Registry) Get(connectorType string) (Connector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.connectors[connectorType]
	if !ok {
		return nil, fmt.Errorf("connector type %q not registered", connectorType)
	}
	return c, nil
}

// Types returns all registered connector type identifiers.
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.connectors))
	for t := range r.connectors {
		types = append(types, t)
	}
	return types
}
