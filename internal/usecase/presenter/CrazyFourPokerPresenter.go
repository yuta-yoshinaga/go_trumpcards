//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CrazyFourPokerPresenter クレイジー 4 ポーカープレゼンターインタフェース
type CrazyFourPokerPresenter interface {
	GamePresenter[interfaces.CrazyFourPokerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.CrazyFourPokerGame) string
}
