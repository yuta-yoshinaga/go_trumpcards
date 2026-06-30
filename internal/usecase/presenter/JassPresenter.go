//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// JassPresenter ヤス(シーバー)プレゼンターインタフェース
type JassPresenter interface {
	GamePresenter[interfaces.JassGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.JassGame) string
}
