//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PaiGowPresenter パイガオポーカープレゼンターインタフェース
type PaiGowPresenter interface {
	GamePresenter[interfaces.PaiGowGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.PaiGowGame) string
}
