package generator

import (
	"fmt"
	"sync"
)

// Generator is the interface that all data generators must implement
type Generator interface {
	Generate(ctx *Context) (interface{}, error)
	Name() string
}

// Registry stores and manages generators
type Registry struct {
	mu         sync.RWMutex
	generators map[string]Generator
}

var (
	defaultRegistry *Registry
	once            sync.Once
)

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]Generator),
	}
}

// DefaultRegistry returns the global default registry
func DefaultRegistry() *Registry {
	once.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}

// Register adds a generator to the registry
func (r *Registry) Register(name string, gen Generator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.generators[name]; exists {
		return fmt.Errorf("generator %s is already registered", name)
	}

	r.generators[name] = gen
	return nil
}

// Get retrieves a generator by name
func (r *Registry) Get(name string) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gen, exists := r.generators[name]
	if !exists {
		return nil, fmt.Errorf("generator %s not found", name)
	}

	return gen, nil
}

// Has checks if a generator exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.generators[name]
	return exists
}

// List returns all registered generator names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}

	return names
}