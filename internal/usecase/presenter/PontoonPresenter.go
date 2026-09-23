//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PontoonPresenter ポンツーン プレゼンターインタフェース
type PontoonPresenter interface {
	GamePresenter[interfaces.PontoonGame]
}
