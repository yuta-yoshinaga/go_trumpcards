//go:build test

package domain

import "testing"

func TestRestoreGamePlayer(t *testing.T) {
	t.Run("nil creates non-human player", func(t *testing.T) {
		got := restoreGamePlayer(nil)
		if got == nil {
			t.Fatal("restoreGamePlayer(nil) = nil")
		}
		if got.GetIsHuman() {
			t.Fatal("restored player is human")
		}
	})

	t.Run("non-nil returns same player", func(t *testing.T) {
		want := NewGamePlayer(true)
		if got := restoreGamePlayer(want); got != want {
			t.Fatal("restoreGamePlayer did not return the original player")
		}
	})
}

func TestAssignIfSet(t *testing.T) {
	t.Run("nil leaves destination unchanged", func(t *testing.T) {
		got := 42
		assignIfSet(&got, (*int)(nil))
		if got != 42 {
			t.Fatalf("destination = %d, want 42", got)
		}
	})

	t.Run("non-nil copies source", func(t *testing.T) {
		got, src := 42, 7
		assignIfSet(&got, &src)
		if got != 7 {
			t.Fatalf("destination = %d, want 7", got)
		}
	})
}
