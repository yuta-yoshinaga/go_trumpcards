//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BarbuPresenter はバルブプレゼンターインタフェース。
type BarbuPresenter interface {
	GamePresenter[interfaces.BarbuGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BarbuGame) string
}
