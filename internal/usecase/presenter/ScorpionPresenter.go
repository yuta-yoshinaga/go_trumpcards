//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ScorpionPresenter スコーピオンプレゼンターインタフェース
type ScorpionPresenter interface {
	GamePresenter[interfaces.ScorpionGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.ScorpionGame) string
	// LegalMovesOutput 指定列のトップカードの合法な移動先を出力する
	LegalMovesOutput(s interfaces.ScorpionGame, col int) string
}
