// /di/di_test.go
package di_test

import (
	"testing"

	"github.com/goletan/di"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// MockService is a simple mock to register in the DI container.
type MockService struct {
	name string
}

// ServiceA depends on ServiceB
type ServiceA struct {
	ServiceB *ServiceB
}

// ServiceB depends on ServiceA
type ServiceB struct {
	ServiceA *ServiceA
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
	mockService := &MockService{name: "TestService"}
	container.Register("mockService", mockService)

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

// TestCircularDependency tests manual handling of circular dependencies.
func TestCircularDependency(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	serviceA := &ServiceA{}
	serviceB := &ServiceB{}

	// Register both services
	container.Register("serviceA", serviceA)
	container.Register("serviceB", serviceB)

	// Trying to resolve circular dependencies manually (since our DI does not handle wiring these)
	// Normally, DI frameworks that support circular dependencies would handle this.
	resolvedA, err := container.Resolve("serviceA")
	if err != nil {
		t.Fatalf("Failed to resolve ServiceA: %v", err)
	}

	// Simulate circular assignment
	resolvedA.(*ServiceA).ServiceB = serviceB
	serviceB.ServiceA = resolvedA.(*ServiceA)

	if resolvedA.(*ServiceA).ServiceB == nil {
		t.Errorf("Expected ServiceB to be assigned to ServiceA, got nil")
	}
	if serviceB.ServiceA == nil {
		t.Errorf("Expected ServiceA to be assigned to ServiceB, got nil")
	}
}

// TestSingletonRegistration tests the registration and resolution of singleton services in the DI container.
func TestSingletonRegistration(t *testing.T) {
	core, _ := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	container := di.NewContainer(logger)

	// Registering a singleton service
	container.RegisterSingleton("singletonService", func() interface{} {
		return &MockService{name: "SingletonService"}
	})

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
	container.RegisterTransient("transientService", func() interface{} {
		return &MockService{name: "TransientService"}
	})

	// Resolving the transient service multiple times
	resolved1, err1 := container.ResolveTransient("transientService")
	if err1 != nil {
		t.Fatalf("Expected transient service to be resolved, got error: %v", err1)
	}
	resolved2, err2 := container.ResolveTransient("transientService")
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

	container.RegisterSingleton("hookedService", func() interface{} {
		return &MockService{name: "HookedService"}
	})

	container.RegisterPreInit("hookedService", func() {
		preInitCalled = true
	})

	container.RegisterPostDestroy("hookedService", func() {
		postDestroyCalled = true
	})

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
