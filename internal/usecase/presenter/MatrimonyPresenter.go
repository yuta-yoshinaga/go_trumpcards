//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MatrimonyPresenter マトリモニー プレゼンターインタフェース
type MatrimonyPresenter interface {
	GamePresenter[interfaces.MatrimonyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.MatrimonyGame) string
}
