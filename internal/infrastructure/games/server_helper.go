//go:build !js || !wasm

package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// BindWebControllerFor is the one-call HTTP-server registration helper every
// game uses. It replaces the old 5-line double-closure with a single call:
// pass the interactor factory and the NewXxxWebController constructor, and
// the helper supplies the outer `func() WebController` closure. Go infers
// the I/P/O type parameters from newCtrl, so call sites stay untyped.
//
// This helper is the DRY/SSoT answer to issue #1458 for the HTTP server.
func BindWebControllerFor[I any, P controller.WebInput, O any](
	name string,
	newInteractor func() I,
	newCtrl func(func() I) *controller.GameWebController[I, P, O],
) {
	BindWebController(name, func() WebController {
		return newCtrl(newInteractor)
	})
}
