//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LaughAndLieDownPresenter ラフ・アンド・ライダウンプレゼンターインタフェース
type LaughAndLieDownPresenter interface {
	GamePresenter[interfaces.LaughAndLieDownGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.LaughAndLieDownGame) string
}
