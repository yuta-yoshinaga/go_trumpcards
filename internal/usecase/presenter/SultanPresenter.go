//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SultanPresenter スルタンプレゼンターインタフェース
type SultanPresenter interface {
	GamePresenter[interfaces.SultanGame]
	// HintOutput ヒント情報を出力する
	HintOutput(su interfaces.SultanGame) string
}
