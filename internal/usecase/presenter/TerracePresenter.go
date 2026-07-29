//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TerracePresenter テラス プレゼンターインタフェース
type TerracePresenter interface {
	GamePresenter[interfaces.TerraceGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TerraceGame) string
}
