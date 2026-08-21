//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NarcoticPresenter ナルコティックプレゼンターインタフェース
type NarcoticPresenter interface {
	GamePresenter[interfaces.NarcoticGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.NarcoticGame) string
}
