//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TarocchiniPresenter タロッキーニのプレゼンターインタフェース
type TarocchiniPresenter interface {
	GamePresenter[interfaces.TarocchiniGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TarocchiniGame) string
}
