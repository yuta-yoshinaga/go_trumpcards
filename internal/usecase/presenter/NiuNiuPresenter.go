//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NiuNiuPresenter 闘牛 プレゼンターインタフェース
type NiuNiuPresenter interface {
	GamePresenter[interfaces.NiuNiuGame]
}
