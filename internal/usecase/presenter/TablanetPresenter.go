//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TablanetPresenter はタブラネット (Tablanet) のプレゼンターインタフェース。
type TablanetPresenter interface {
	GamePresenter[interfaces.TablanetGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TablanetGame) string
}
