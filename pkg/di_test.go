// /di/pkg/di_test.go
package di_test

import (
	"testing"

	di "github.com/goletan/di/pkg"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// MockService is a simple mock to register in the DI container.
type MockService struct {
	name string
}

func (ms *MockService) Name() string {
	return ms.name
}

// TestContainerRegistration tests the registration and resolution of services in the DI container.
func TestContainerRegistration(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	// Registering a mock service
	mockService := func() interface{} {
		return &MockService{name: "TestService"}
	}
	container.Register("mockService", mockService, di.Transient)

	// Resolving the registered service
	resolved, err := container.Resolve("mockService")
	if err != nil {
		t.Fatalf("Expected service to be resolved, got error: %v", err)
	}

	// Type assertion to validate resolved type
	if resolved.(*MockService).Name() != "TestService" {
		t.Errorf("Expected service name 'TestService', got '%s'", resolved.(*MockService).Name())
	}
}

// TestResolveUnregisteredService tests resolution of an unregistered service.
func TestResolveUnregisteredService(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	_, err := container.Resolve("unregisteredService")
	if err == nil {
		t.Fatalf("Expected error when resolving unregistered service, got nil")
	}
}

// TestSingletonRegistration tests the registration and resolution of singleton services in the DI container.
func TestSingletonRegistration(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	// Registering a singleton service
	container.Register("singletonService", func() interface{} {
		return &MockService{name: "SingletonService"}
	}, di.Singleton)

	// Resolving the singleton service multiple times
	resolved1, err1 := container.Resolve("singletonService")
	if err1 != nil {
		t.Fatalf("Expected singleton service to be resolved, got error: %v", err1)
	}
	resolved2, err2 := container.Resolve("singletonService")
	if err2 != nil {
		t.Fatalf("Expected singleton service to be resolved, got error: %v", err2)
	}

	// Validate that the resolved instances are the same
	if resolved1 != resolved2 {
		t.Errorf("Expected the same instance for singleton service, got different instances")
	}
}

// TestTransientRegistration tests the registration and resolution of transient services in the DI container.
func TestTransientRegistration(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	// Registering a transient service
	container.Register("transientService", func() interface{} {
		return &MockService{name: "TransientService"}
	}, di.Transient)

	// Resolving the transient service multiple times
	resolved1, err1 := container.Resolve("transientService")
	if err1 != nil {
		t.Fatalf("Expected transient service to be resolved, got error: %v", err1)
	}
	resolved2, err2 := container.Resolve("transientService")
	if err2 != nil {
		t.Fatalf("Expected transient service to be resolved, got error: %v", err2)
	}

	// Validate that the resolved instances are different
	if resolved1 == resolved2 {
		t.Errorf("Expected different instances for transient service, got the same instance")
	}
}

// TestLifecycleHooks tests the pre-initialization and post-destroy lifecycle hooks.
func TestLifecycleHooks(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	preInitCalled := false
	postDestroyCalled := false

	// Register a service with pre-initialization and post-destroy hooks
	container.Register(
		"hookedService",
		func() interface{} {
			return &MockService{name: "HookedService"}
		},
		di.Singleton,
		di.WithInitHook(func() {
			preInitCalled = true
		}),
		di.WithDestroyHook(func() {
			postDestroyCalled = true
		}),
	)

	// Resolve the service to trigger pre-initialization
	_, err := container.Resolve("hookedService")
	if err != nil {
		t.Fatalf("Expected hooked service to be resolved, got error: %v", err)
	}

	if !preInitCalled {
		t.Errorf("Expected pre-initialization hook to be called, but it wasn't")
	}

	// Destroy the service to trigger post-destroy hook
	container.Destroy("hookedService")

	if !postDestroyCalled {
		t.Errorf("Expected post-destroy hook to be called, but it wasn't")
	}
}
