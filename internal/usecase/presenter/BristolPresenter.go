//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BristolPresenter ブリストルプレゼンターインタフェース
type BristolPresenter interface {
	GamePresenter[interfaces.BristolGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BristolGame) string
}
