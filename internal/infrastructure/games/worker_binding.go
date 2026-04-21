//go:build js && wasm

package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/infrastructure/worker"
)

// WorkerBinding captures the ingredients needed to wire a game into a Cloudflare Worker.
type WorkerBinding[I worker.SnapshotIF, P controller.WebInput, O any] struct {
	Name     string
	Category Category

	// NewInteractor produces a fresh interactor bound to a fresh presenter
	// (used when no prior KV state exists for this session).
	NewInteractor func() I

	// RestoreInteractor rehydrates an interactor from serialised KV state.
	RestoreInteractor func(data []byte) (I, error)

	// NewControllerWithProvider is the NewXxxWebControllerWithProvider
	// (KV-session) variant the worker uses to inject a session provider.
	NewControllerWithProvider func(
		controller.SessionProvider[I], func() I,
	) *controller.GameWebController[I, P, O]
}

// Bind registers this binding via RegisterKVGame. Panics via BindWorker if
// Name is unknown or Category does not match the registry.
func (b WorkerBinding[I, P, O]) Bind() {
	RegisterKVGame(b.Name, b.Category,
		b.NewInteractor, b.RestoreInteractor, b.NewControllerWithProvider)
}
