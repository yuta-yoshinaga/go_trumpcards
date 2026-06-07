//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TriPeaksPresenter トリピークスプレゼンターインタフェース
type TriPeaksPresenter interface {
	GamePresenter[interfaces.TriPeaksGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TriPeaksGame) string
}
