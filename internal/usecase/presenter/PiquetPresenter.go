//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PiquetPresenter Piquetプレゼンターインタフェース
type PiquetPresenter interface {
	GamePresenter[interfaces.PiquetGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PiquetGame) string
}
