//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SchnapsenPresenter シュナプセンプレゼンターインタフェース
type SchnapsenPresenter interface {
	GamePresenter[interfaces.SchnapsenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SchnapsenGame) string
}
