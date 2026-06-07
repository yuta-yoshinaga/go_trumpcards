//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EasthavenPresenter イーストヘイブンプレゼンターインタフェース
type EasthavenPresenter interface {
	GamePresenter[interfaces.EasthavenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(e interfaces.EasthavenGame) string
}
