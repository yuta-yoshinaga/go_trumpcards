//go:build !js || !wasm

package games

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
)

// BindWebControllerFor is a thin generic wrapper over BindWebController that
// removes the double-closure boilerplate shared by every game. Callers pass
// the interactor factory and the matching NewXxxWebController constructor;
// the outer `func() WebController` closure is supplied here.
func BindWebControllerFor[I any, P controller.WebInput, O any](
	name string,
	newInteractor func() I,
	newCtrl func(func() I) *controller.GameWebController[I, P, O],
) {
	BindWebController(name, func() WebController {
		return newCtrl(newInteractor)
	})
}
