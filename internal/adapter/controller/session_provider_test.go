package controller_test

import (
	"strings"
	"testing"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

func TestMemorySessionProvider_AcquireAndRelease(t *testing.T) {
	p := controller.NewMemorySessionProvider[int]()
	defer p.Stop()
	calls := 0
	factory := func() int { calls++; return calls }

	val, release, ok := p.Acquire("s1", factory)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if val != 1 {
		t.Errorf("want 1, got %d", val)
	}
	release()

	// Same session returns same value without calling factory again.
	val2, release2, ok2 := p.Acquire("s1", factory)
	if !ok2 {
		t.Fatal("expected ok=true on second Acquire")
	}
	if val2 != 1 {
		t.Errorf("want 1 on reuse, got %d", val2)
	}
	if calls != 1 {
		t.Errorf("factory called %d times, want 1", calls)
	}
	release2()
}

func TestMemorySessionProvider_TooLongID(t *testing.T) {
	p := controller.NewMemorySessionProvider[int]()
	defer p.Stop()
	longID := strings.Repeat("x", controller.SessionMaxIDLen+1)

	_, release, ok := p.Acquire(longID, func() int { return 1 })
	if ok {
		t.Fatal("expected ok=false for too-long session ID")
	}
	if release != nil {
		t.Error("expected nil release for rejected session")
	}
}

func TestMemorySessionProvider_Store(t *testing.T) {
	p := controller.NewMemorySessionProvider[int]()
	defer p.Stop()

	if p.Store() == nil {
		t.Fatal("Store() should return the underlying SessionStore")
	}
}
