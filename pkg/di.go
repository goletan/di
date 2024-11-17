// /di/pkg/di.go
package di

import (
	"github.com/goletan/di/internal/container"
	"go.uber.org/zap"
)

type Container struct {
	internal *container.Container
}

// LifetimeType defines the type of lifetime a service should have in the DI container.
type LifetimeType = container.LifetimeType

const (
	// LifetimeSingleton defines that the service should be a singleton.
	LifetimeSingleton LifetimeType = container.LifetimeSingleton
	// LifetimeTransient defines that the service should be created every time it is requested.
	LifetimeTransient LifetimeType = container.LifetimeTransient
	// LifetimeScoped defines that the service should be created once per scope (not yet implemented).
	LifetimeScoped LifetimeType = container.LifetimeScoped
)

// NewContainer creates a new DI container with the public API.
func NewContainer(logger *zap.Logger) *Container {
	return &Container{
		internal: container.NewContainer(logger),
	}
}

// Register adds a new service to the DI container.
func (c *Container) Register(name string, constructor func() interface{}, lifetime LifetimeType) {
	c.internal.Register(name, constructor, lifetime)
}

// Resolve retrieves a service by name from the DI container.
func (c *Container) Resolve(name string) (interface{}, error) {
	return c.internal.Resolve(name)
}

// MustResolve retrieves a service and panics if not found, useful for essential services.
func (c *Container) MustResolve(name string) interface{} {
	return c.internal.MustResolve(name)
}

// Destroy removes a service from the container.
func (c *Container) Destroy(name string) {
	c.internal.Destroy(name)
}
