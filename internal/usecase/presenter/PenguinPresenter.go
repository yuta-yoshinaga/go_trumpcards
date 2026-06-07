//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PenguinPresenter ペンギンプレゼンターインタフェース
type PenguinPresenter interface {
	GamePresenter[interfaces.PenguinGame]
	// HintOutput ヒント情報を出力する
	HintOutput(p interfaces.PenguinGame) string
}
