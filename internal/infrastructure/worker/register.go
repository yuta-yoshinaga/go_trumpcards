//go:build js && wasm

// Package worker provides shared helpers for Cloudflare Worker entry points.
package worker

import (
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// SnapshotIF is the interface required on all interactors for KV persistence.
type SnapshotIF interface {
	Snapshot() ([]byte, error)
}

// RegisterKV creates a KV-backed session provider for a game and registers the
// controller on the given mux. Returns an error if the KV namespace cannot be
// initialised; callers should handle it (typically with log.Fatal in main).
//
// I must implement SnapshotIF; this is enforced at compile time.
// newCtrl is a controller constructor such as controller.NewBlackJackWebControllerWithProvider.
// Passing the constructor directly avoids the 4-line closure wrapper that was
// previously required at every call site.
func RegisterKV[I SnapshotIF, P controller.WebInput, O any](
	mux *http.ServeMux,
	path string,
	kvPrefix string,
	factory func() I,
	restore func([]byte) (I, error),
	newCtrl func(controller.SessionProvider[I], func() I) *controller.GameWebController[I, P, O],
) error {
	kvProvider, err := controller.NewKVSessionProvider[I](
		"GAME_SESSIONS", kvPrefix,
		func(i I) ([]byte, error) { return i.Snapshot() },
		restore,
	)
	if err != nil {
		return err
	}
	ctrl := newCtrl(kvProvider, factory)
	mux.HandleFunc(path, ctrl.Exec)
	return nil
}
