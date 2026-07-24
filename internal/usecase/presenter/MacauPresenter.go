//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MacauPresenter マカオプレゼンターインタフェース
type MacauPresenter interface {
	GamePresenter[interfaces.MacauGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MacauGame) string
}
