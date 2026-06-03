//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EightOffPresenter エイトオフプレゼンターインタフェース
type EightOffPresenter interface {
	GamePresenter[interfaces.EightOffGame]
	// HintOutput ヒント情報を出力する
	HintOutput(e interfaces.EightOffGame) string
}
