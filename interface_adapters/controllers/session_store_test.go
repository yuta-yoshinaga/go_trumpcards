package controllers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/interface_adapters/controllers"
)

func TestSessionStore_GetCreatesOnce(t *testing.T) {
	calls := 0
	store := controllers.NewSessionStore[int]()
	defer store.Stop()
	factory := func() int { calls++; return calls }

	v1, ok1 := store.Get("session-a", factory)
	if !ok1 {
		t.Fatal("expected ok=true on first Get")
	}
	if v1 != 1 {
		t.Errorf("want value 1, got %d", v1)
	}

	// Second Get for same session must reuse — factory must NOT be called again.
	v2, ok2 := store.Get("session-a", factory)
	if !ok2 {
		t.Fatal("expected ok=true on second Get")
	}
	if v2 != 1 {
		t.Errorf("want same value 1 on reuse, got %d", v2)
	}
	if calls != 1 {
		t.Errorf("factory called %d times, want exactly 1", calls)
	}
}

func TestSessionStore_DifferentSessionsAreIsolated(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	defer store.Stop()
	counter := 0
	factory := func() int { counter++; return counter }

	v1, _ := store.Get("session-A", factory)
	v2, _ := store.Get("session-B", factory)
	if v1 == v2 {
		t.Errorf("different sessions should get different values, both got %d", v1)
	}
	if store.Len() != 2 {
		t.Errorf("want 2 sessions, got %d", store.Len())
	}
}

func TestSessionStore_TooLongID(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	defer store.Stop()
	longID := strings.Repeat("x", controllers.SessionMaxIDLen+1)
	_, ok := store.Get(longID, func() int { return 1 })
	if ok {
		t.Fatal("expected ok=false for sessionId longer than SessionMaxIDLen")
	}
	if store.Len() != 0 {
		t.Errorf("store should be empty after rejected insert, got Len=%d", store.Len())
	}
}

func TestSessionStore_ExactMaxIDLen(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	defer store.Stop()
	exactID := strings.Repeat("x", controllers.SessionMaxIDLen)
	_, ok := store.Get(exactID, func() int { return 1 })
	if !ok {
		t.Fatal("expected ok=true for sessionId of exactly SessionMaxIDLen characters")
	}
}

func TestSessionStore_MaxCount(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	defer store.Stop()

	for i := 0; i < controllers.SessionMaxCount; i++ {
		id := fmt.Sprintf("s%d", i)
		_, ok := store.Get(id, func() int { return i })
		if !ok {
			t.Fatalf("expected ok=true at i=%d (before limit)", i)
		}
	}
	if store.Len() != controllers.SessionMaxCount {
		t.Errorf("want Len=%d, got %d", controllers.SessionMaxCount, store.Len())
	}

	// One more should fail.
	_, ok := store.Get("overflow", func() int { return 0 })
	if ok {
		t.Fatal("expected ok=false when store is at capacity")
	}
}

func TestSessionStore_EvictExpiredDoesNotRemoveActiveSession(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	defer store.Stop()

	store.Get("live", func() int { return 42 })
	if store.Len() != 1 {
		t.Fatalf("expected 1 session, got %d", store.Len())
	}

	// Eviction should NOT remove a session that was just accessed.
	store.EvictExpired()
	if store.Len() != 1 {
		t.Errorf("expected session to survive EvictExpired (TTL not reached), got Len=%d", store.Len())
	}
}

func TestSessionStore_StopIsIdempotent(t *testing.T) {
	store := controllers.NewSessionStore[int]()
	store.Stop()
	store.Stop() // calling Stop twice must not panic
}
