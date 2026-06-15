//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NapPresenter ナップのプレゼンターインタフェース
type NapPresenter interface {
	GamePresenter[interfaces.NapGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.NapGame) string
}
