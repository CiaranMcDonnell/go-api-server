package cache

import "time"

// Store is the interface that both in-memory Cache and future Redis
// implementations satisfy. Callers can depend on this interface to
// swap backends without changing application code.
type Store[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
	Delete(key K)
}
