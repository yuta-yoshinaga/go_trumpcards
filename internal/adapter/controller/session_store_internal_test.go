package controller

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
