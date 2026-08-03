//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MinchiatePresenter ミンキアーテのプレゼンターインタフェース
type MinchiatePresenter interface {
	GamePresenter[interfaces.MinchiateGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MinchiateGame) string
}
