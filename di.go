// /di/di.go

package di

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Container manages the dependencies and lifecycle of services.
type Container struct {
	services    map[string]interface{}
	singletons  map[string]func() interface{}
	instances   map[string]interface{}
	preInit     map[string]func()
	postDestroy map[string]func()
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewContainer creates a new DI container.
func NewContainer(logger *zap.Logger) *Container {
	return &Container{
		services:    make(map[string]interface{}),
		singletons:  make(map[string]func() interface{}),
		instances:   make(map[string]interface{}),
		preInit:     make(map[string]func()),
		postDestroy: make(map[string]func()),
		logger:      logger,
	}
}

// Register adds a new service to the DI container.
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
	c.logger.Info("Service registered", zap.String("service_name", name))
}

// RegisterSingleton registers a service as a singleton, ensuring only one instance is used.
func (c *Container) RegisterSingleton(name string, constructor func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singletons[name] = constructor
	c.logger.Info("Singleton service registered", zap.String("service_name", name))
}

// RegisterPreInit registers a function that will be called before initializing a service.
func (c *Container) RegisterPreInit(name string, fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.preInit[name] = fn
	c.logger.Info("Pre-initialization hook registered", zap.String("service_name", name))
}

// RegisterPostDestroy registers a function that will be called after destroying a service.
func (c *Container) RegisterPostDestroy(name string, fn func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.postDestroy[name] = fn
	c.logger.Info("Post-destroy hook registered", zap.String("service_name", name))
}

// Resolve retrieves a service by name from the DI container.
func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if the service is registered as a singleton and initialize it if needed
	if constructor, exists := c.singletons[name]; exists {
		if instance, found := c.instances[name]; found {
			return instance, nil
		}
		// Upgrade lock to initialize the singleton instance
		c.mu.RUnlock()
		c.mu.Lock()
		defer c.mu.Unlock()
		if instance, found := c.instances[name]; found {
			// Check again to avoid race conditions
			return instance, nil
		}
		if preInit, exists := c.preInit[name]; exists {
			preInit()
		}
		instance := constructor()
		c.instances[name] = instance
		c.logger.Info("Singleton service initialized", zap.String("service_name", name))
		return instance, nil
	}

	// Check if the service is registered as a regular service
	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}
	return service, nil
}

// MustResolve retrieves a service and panics if not found, useful for essential services.
func (c *Container) MustResolve(name string) interface{} {
	service, err := c.Resolve(name)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve service: %s", err))
	}
	return service
}

// Destroy removes a service from the container and calls the post-destroy hook if available.
func (c *Container) Destroy(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.services[name]; exists {
		delete(c.services, name)
		c.logger.Info("Service destroyed", zap.String("service_name", name))
	}
	if _, exists := c.instances[name]; exists {
		delete(c.instances, name)
		if postDestroy, exists := c.postDestroy[name]; exists {
			postDestroy()
		}
		c.logger.Info("Singleton instance destroyed", zap.String("service_name", name))
	}
}

// RegisterTransient registers a transient service, always providing a new instance.
func (c *Container) RegisterTransient(name string, constructor func() interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = constructor
	c.logger.Info("Transient service registered", zap.String("service_name", name))
}

// ResolveTransient retrieves a new instance of a transient service by name.
func (c *Container) ResolveTransient(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	constructor, ok := service.(func() interface{})
	if !ok {
		return nil, fmt.Errorf("service %s is not a transient constructor", name)
	}
	return constructor(), nil
}
