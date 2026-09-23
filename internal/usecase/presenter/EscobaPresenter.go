//go:build !js || !wasm || extra7

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EscobaPresenter エスコバプレゼンターインタフェース。
type EscobaPresenter interface {
	GamePresenter[interfaces.EscobaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.EscobaGame) string
}
