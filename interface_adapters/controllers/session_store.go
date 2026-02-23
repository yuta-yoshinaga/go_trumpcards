package controllers

import (
	"sync"
	"time"
)

const (
	// SessionMaxIDLen is the maximum allowed length of a session ID.
	SessionMaxIDLen = 256
	// SessionMaxCount is the maximum number of concurrent sessions.
	SessionMaxCount = 10000
	// SessionTTL is the inactivity duration after which a session is evicted.
	SessionTTL = 1 * time.Hour
	// sessionCleanupInterval is how often the background goroutine runs eviction.
	sessionCleanupInterval = 10 * time.Minute
)

type sessionEntry[T any] struct {
	value    T
	mu       sync.Mutex
	lastUsed time.Time
}

// SessionStore is a generic, goroutine-safe map of session ID → value with
// TTL-based eviction and a hard cap on concurrent sessions.
type SessionStore[T any] struct {
	mu      sync.Mutex
	entries map[string]*sessionEntry[T]
	stopCh  chan struct{}
	once    sync.Once
}

// NewSessionStore creates a SessionStore and starts a background cleanup goroutine.
// Call Stop() when the store is no longer needed.
func NewSessionStore[T any]() *SessionStore[T] {
	s := &SessionStore[T]{
		entries: make(map[string]*sessionEntry[T]),
		stopCh:  make(chan struct{}),
	}
	go s.cleanupLoop()
	return s
}

// Get returns the value for the given sessionId, creating it via factory if it
// does not yet exist. The second return value is false when the sessionId is
// invalid (too long) or the store is at capacity.
func (s *SessionStore[T]) Get(id string, factory func() T) (T, bool) {
	var zero T
	if len(id) > SessionMaxIDLen {
		return zero, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry, ok := s.entries[id]; ok {
		entry.lastUsed = time.Now()
		return entry.value, true
	}
	if len(s.entries) >= SessionMaxCount {
		return zero, false
	}
	val := factory()
	s.entries[id] = &sessionEntry[T]{value: val, lastUsed: time.Now()}
	return val, true
}

// GetWithLock returns the value for the given sessionId (creating it via factory
// if needed) along with a per-session mutex. The caller must use
// defer mu.Unlock() after mu.Lock() to ensure the lock is always released.
// This prevents concurrent requests for the same session from racing on
// shared state.
func (s *SessionStore[T]) GetWithLock(id string, factory func() T) (T, *sync.Mutex, bool) {
	var zero T
	if len(id) > SessionMaxIDLen {
		return zero, nil, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[id]
	if ok {
		entry.lastUsed = time.Now()
		return entry.value, &entry.mu, true
	}
	if len(s.entries) >= SessionMaxCount {
		return zero, nil, false
	}
	entry = &sessionEntry[T]{value: factory(), lastUsed: time.Now()}
	s.entries[id] = entry
	return entry.value, &entry.mu, true
}

// EvictExpired removes sessions that have not been used within SessionTTL.
// It is exported so tests can trigger eviction without waiting for the ticker.
func (s *SessionStore[T]) EvictExpired() {
	cutoff := time.Now().Add(-SessionTTL)
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, entry := range s.entries {
		if entry.lastUsed.Before(cutoff) {
			delete(s.entries, id)
		}
	}
}

// Len returns the current number of sessions.
func (s *SessionStore[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}

// Stop shuts down the background cleanup goroutine. Safe to call multiple times.
func (s *SessionStore[T]) Stop() {
	s.once.Do(func() { close(s.stopCh) })
}

func (s *SessionStore[T]) cleanupLoop() {
	ticker := time.NewTicker(sessionCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.EvictExpired()
		case <-s.stopCh:
			return
		}
	}
}
