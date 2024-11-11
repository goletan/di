// /di/internal/container/lifetime.go
package pkg

// Lifetime defines the lifetime of a registered service.
type Lifetime int

const (
	Singleton Lifetime = iota
	Transient
)

func (l Lifetime) String() string {
	switch l {
	case Singleton:
		return "Singleton"
	case Transient:
		return "Transient"
	default:
		return "Unknown"
	}
}
