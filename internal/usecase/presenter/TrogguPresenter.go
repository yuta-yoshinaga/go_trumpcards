//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TrogguPresenter トロッグのプレゼンターインタフェース。
type TrogguPresenter interface {
	GamePresenter[interfaces.TrogguGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TrogguGame) string
}
