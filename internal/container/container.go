package container

import (
	"errors"
	"sync"

	"github.com/goletan/observability/shared/logger"
	"go.uber.org/zap"
)

type Container struct {
	services map[string]*serviceEntry
	logger   *logger.ZapLogger
	mu       sync.RWMutex
}

type serviceEntry struct {
	constructor func() interface{}
	instance    interface{}
	lifetime    LifetimeType
}

func NewContainer(log *logger.ZapLogger) *Container {
	return &Container{
		services: make(map[string]*serviceEntry),
		logger:   log,
	}
}

func (c *Container) Register(name string, constructor func() interface{}, lifetime LifetimeType) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = &serviceEntry{
		constructor: constructor,
		lifetime:    lifetime,
	}
	c.logger.Info("Service registered", zap.String("service_name", name), zap.String("lifetime", lifetime.String()))
}

func (c *Container) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	entry, exists := c.services[name]
	c.mu.RUnlock()

	if !exists {
		c.logger.Error("Service not found", zap.String("service_name", name))
		return nil, errors.New("service not found: " + name)
	}

	if entry.lifetime == LifetimeSingleton && entry.instance != nil {
		return entry.instance, nil
	}

	instance := entry.constructor()
	if entry.lifetime == LifetimeSingleton {
		entry.instance = instance
	}

	return instance, nil
}

func (c *Container) MustResolve(name string) interface{} {
	instance, err := c.Resolve(name)
	if err != nil {
		panic(err)
	}
	return instance
}

func (c *Container) Destroy(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.services[name]; exists {
		delete(c.services, name)
		c.logger.Info("Service destroyed", zap.String("service_name", name))
	}
}
