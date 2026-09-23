//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SetteEMezzoPresenter セッテ・エ・メッツォ プレゼンターインタフェース
type SetteEMezzoPresenter interface {
	GamePresenter[interfaces.SetteEMezzoGame]
}
