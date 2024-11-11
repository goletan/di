// /di/internal/container/container.go

package container

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Container manages the dependencies and lifecycle of services.
type Container struct {
	services  map[string]serviceDefinition
	instances map[string]interface{}
	mu        sync.RWMutex
	logger    *zap.Logger
}

type serviceDefinition struct {
	constructor func() interface{}
	lifetime    Lifetime
	initHook    func()
	destroyHook func()
}

// NewContainer creates a new DI container.
func NewContainer(logger *zap.Logger) *Container {
	return &Container{
		services:  make(map[string]serviceDefinition),
		instances: make(map[string]interface{}),
		logger:    logger,
	}
}

// Register adds a new service to the DI container with a specified lifetime.
func (c *Container) Register(name string, constructor func() interface{}, lifetime Lifetime, opts ...ServiceOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceDef := serviceDefinition{
		constructor: constructor,
		lifetime:    lifetime,
	}

	for _, opt := range opts {
		opt(&serviceDef)
	}

	c.services[name] = serviceDef
	c.logger.Info("Service registered", zap.String("service_name", name), zap.String("lifetime", lifetime.String()))
}

// ServiceOption is a function that configures a service definition.
type ServiceOption func(*serviceDefinition)

// WithInitHook adds an initialization hook to a service.
func WithInitHook(hook func()) ServiceOption {
	return func(s *serviceDefinition) {
		s.initHook = hook
	}
}

// WithDestroyHook adds a destroy hook to a service.
func WithDestroyHook(hook func()) ServiceOption {
	return func(s *serviceDefinition) {
		s.destroyHook = hook
	}
}

// Resolve retrieves a service by name from the DI container.
func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	serviceDef, exists := c.services[name]
	c.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	if serviceDef.lifetime == Singleton {
		c.mu.Lock()
		defer c.mu.Unlock()

		if instance, found := c.instances[name]; found {
			return instance, nil
		}

		if serviceDef.initHook != nil {
			serviceDef.initHook()
		}
		instance := serviceDef.constructor()
		c.instances[name] = instance
		c.logger.Info("Singleton service initialized", zap.String("service_name", name))
		return instance, nil
	}

	instance := serviceDef.constructor()
	if serviceDef.initHook != nil {
		serviceDef.initHook()
	}
	c.logger.Info("Transient service resolved", zap.String("service_name", name))
	return instance, nil
}

// MustResolve retrieves a service and panics if not found, useful for essential services.
func (c *Container) MustResolve(name string) interface{} {
	service, err := c.Resolve(name)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve service: %s", err))
	}
	return service
}

// Destroy removes a service from the container and calls the destroy hook if available.
func (c *Container) Destroy(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	serviceDef, exists := c.services[name]
	if !exists {
		c.logger.Warn("Service not found during destroy", zap.String("service_name", name))
		return
	}

	if _, found := c.instances[name]; found {
		if serviceDef.destroyHook != nil {
			serviceDef.destroyHook()
		}
		delete(c.instances, name)
		c.logger.Info("Singleton instance destroyed", zap.String("service_name", name))
	}

	delete(c.services, name)
	c.logger.Info("Service destroyed", zap.String("service_name", name))
}
