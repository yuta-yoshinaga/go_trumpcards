package controller

// SessionRelease is a callback invoked after a request finishes processing.
// For in-memory sessions it unlocks the per-session mutex; for KV-backed
// sessions it serialises the interactor state and writes it back to KV.
type SessionRelease func()

// SessionProvider is the interface that abstracts session storage.
// In-memory and KV-backed implementations both satisfy this contract so
// that GameWebController works identically on Docker and Workers.
type SessionProvider[T any] interface {
	// Acquire returns the interactor for the given session ID (creating it
	// via factory when the session does not exist), a release function that
	// MUST be called when the request is done, and a success flag.
	Acquire(id string, factory func() T) (T, SessionRelease, bool)
	// Stop releases background resources owned by the provider.
	Stop()
}

// MemorySessionProvider wraps SessionStore to satisfy SessionProvider.
type MemorySessionProvider[T any] struct {
	store *SessionStore[T]
}

// NewMemorySessionProvider creates a MemorySessionProvider backed by an
// in-memory SessionStore with background TTL eviction.
func NewMemorySessionProvider[T any]() *MemorySessionProvider[T] {
	return &MemorySessionProvider[T]{store: NewSessionStore[T]()}
}

// Acquire retrieves (or creates) the session value and locks the per-session
// mutex. The returned release function unlocks it.
func (m *MemorySessionProvider[T]) Acquire(id string, factory func() T) (T, SessionRelease, bool) {
	val, mu, ok := m.store.GetWithLock(id, factory)
	if !ok {
		return val, nil, false
	}
	mu.Lock()
	return val, mu.Unlock, true
}

// Stop shuts down the background cleanup goroutine.
func (m *MemorySessionProvider[T]) Stop() {
	m.store.Stop()
}

// Store returns the underlying SessionStore for testing/inspection.
func (m *MemorySessionProvider[T]) Store() *SessionStore[T] {
	return m.store
}
