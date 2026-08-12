//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PigPresenter ピッグプレゼンターインタフェース
type PigPresenter interface {
	GamePresenter[interfaces.PigGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.PigGame) string
}
