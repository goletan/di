package container

type LifetimeType int

const (
	LifetimeSingleton LifetimeType = iota
	LifetimeTransient
	LifetimeScoped
)

func (l LifetimeType) String() string {
	switch l {
	case LifetimeSingleton:
		return "Singleton"
	case LifetimeTransient:
		return "Transient"
	case LifetimeScoped:
		return "Scoped"
	default:
		return "Unknown"
	}
}
