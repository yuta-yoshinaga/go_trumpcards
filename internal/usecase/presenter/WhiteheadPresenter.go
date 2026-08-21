//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WhiteheadPresenter ホワイトヘッドプレゼンターインタフェース
type WhiteheadPresenter interface {
	GamePresenter[interfaces.WhiteheadGame]
	// HintOutput ヒント情報を出力する
	HintOutput(k interfaces.WhiteheadGame) string
}
