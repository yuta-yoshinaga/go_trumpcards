package controller

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGet_DoubleCheck_ExistingEntry exercises the branch at line 69-71 where,
// after releasing the global lock to run factory(), another goroutine has
// already inserted an entry for the same ID.
func TestGet_DoubleCheck_ExistingEntry(t *testing.T) {
	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}
	defer s.Stop()

	// The factory simulates a race: while it runs (lock is released), another
	// goroutine inserts an entry for the same key.
	factory := func() int {
		s.mu.Lock()
		s.entries["race"] = &sessionEntry[int]{value: 42, mu: &sync.Mutex{}, lastUsed: time.Now()}
		s.mu.Unlock()
		return 99
	}

	val, ok := s.Get("race", factory)
	assert.True(t, ok)
	// Should return the value inserted by the "other goroutine" (42), not factory's 99.
	assert.Equal(t, 42, val)
}

// TestGet_DoubleCheck_AtCapacity exercises the branch at line 73-74 where,
// after releasing the global lock to run factory(), other goroutines fill the
// store to capacity.
func TestGet_DoubleCheck_AtCapacity(t *testing.T) {
	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}
	defer s.Stop()

	// Start with SessionMaxCount-1 entries so the first capacity check passes.
	for i := 0; i < SessionMaxCount-1; i++ {
		s.entries[fmt.Sprintf("x%d", i)] = &sessionEntry[int]{value: i, mu: &sync.Mutex{}, lastUsed: time.Now()}
	}

	// factory fills the last slot, so when Get re-acquires the lock for the
	// double-check, the store is at capacity.
	factory := func() int {
		s.mu.Lock()
		s.entries["blocker"] = &sessionEntry[int]{value: -1, mu: &sync.Mutex{}, lastUsed: time.Now()}
		s.mu.Unlock()
		return 999
	}

	val, ok := s.Get("new-key", factory)
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

// TestGetWithLock_DoubleCheck_ExistingEntry exercises the branch at line 106-108.
func TestGetWithLock_DoubleCheck_ExistingEntry(t *testing.T) {
	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}
	defer s.Stop()

	existingMu := &sync.Mutex{}
	factory := func() int {
		s.mu.Lock()
		s.entries["race"] = &sessionEntry[int]{value: 42, mu: existingMu, lastUsed: time.Now()}
		s.mu.Unlock()
		return 99
	}

	val, mu, ok := s.GetWithLock("race", factory)
	assert.True(t, ok)
	assert.Equal(t, 42, val)
	assert.Equal(t, existingMu, mu)
}

// TestGetWithLock_DoubleCheck_AtCapacity exercises the branch at line 110-111.
func TestGetWithLock_DoubleCheck_AtCapacity(t *testing.T) {
	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}
	defer s.Stop()

	for i := 0; i < SessionMaxCount-1; i++ {
		s.entries[fmt.Sprintf("x%d", i)] = &sessionEntry[int]{value: i, mu: &sync.Mutex{}, lastUsed: time.Now()}
	}

	// factory fills the last slot while lock is released
	factory := func() int {
		s.mu.Lock()
		s.entries["blocker"] = &sessionEntry[int]{value: -1, mu: &sync.Mutex{}, lastUsed: time.Now()}
		s.mu.Unlock()
		return 999
	}

	val, mu, ok := s.GetWithLock("new-key", factory)
	assert.False(t, ok)
	assert.Equal(t, 0, val)
	assert.Nil(t, mu)
}

func TestEvictExpired_RemovesStaleSession(t *testing.T) {
	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}
	defer s.Stop()

	s.entries["stale"] = &sessionEntry[int]{
		value:    1,
		mu:       &sync.Mutex{},
		lastUsed: time.Now().Add(-SessionTTL - time.Second),
	}
	s.entries["fresh"] = &sessionEntry[int]{
		value:    2,
		mu:       &sync.Mutex{},
		lastUsed: time.Now(),
	}

	s.EvictExpired()

	assert.Equal(t, 1, len(s.entries))
	_, ok := s.entries["stale"]
	assert.False(t, ok)
	_, ok = s.entries["fresh"]
	assert.True(t, ok)
}

func TestCleanupLoop_EvictsOnTick(t *testing.T) {
	// Temporarily shorten the cleanup interval so the ticker fires quickly.
	orig := sessionCleanupInterval
	sessionCleanupInterval = 10 * time.Millisecond
	defer func() { sessionCleanupInterval = orig }()

	s := &SessionStore[int]{
		entries: make(map[string]*sessionEntry[int]),
		stopCh:  make(chan struct{}),
	}

	s.entries["stale"] = &sessionEntry[int]{
		value:    1,
		mu:       &sync.Mutex{},
		lastUsed: time.Now().Add(-SessionTTL - time.Second),
	}

	go s.cleanupLoop()

	// Wait for the ticker to fire and evict the stale session.
	assert.Eventually(t, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return len(s.entries) == 0
	}, 1*time.Second, 5*time.Millisecond)

	s.Stop()
}
