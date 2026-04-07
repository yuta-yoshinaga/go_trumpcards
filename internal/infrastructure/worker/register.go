//go:build js && wasm

// Package worker provides shared helpers for Cloudflare Worker entry points.
package worker

import (
	"log"
	"net/http"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// RegisterKV creates a KV-backed session provider for a game and registers the
// controller on the given mux. It eliminates the repeated boilerplate of
// creating a provider, building a controller, and wiring up the route.
//
// newCtrl is a controller constructor such as controller.NewBlackJackWebControllerWithProvider.
// Passing the constructor directly avoids the 4-line closure wrapper that was
// previously required at every call site.
func RegisterKV[I any, P controller.WebInput, O any](
	mux *http.ServeMux,
	path string,
	kvPrefix string,
	factory func() I,
	restore func([]byte) (I, error),
	newCtrl func(controller.SessionProvider[I], func() I) *controller.GameWebController[I, P, O],
) {
	kvProvider, err := controller.NewKVSessionProvider[I](
		"GAME_SESSIONS", kvPrefix,
		func(i I) ([]byte, error) {
			return any(i).(interface{ Snapshot() ([]byte, error) }).Snapshot()
		},
		restore,
	)
	if err != nil {
		log.Fatal(err)
	}
	ctrl := newCtrl(kvProvider, factory)
	mux.HandleFunc(path, ctrl.Exec)
}
