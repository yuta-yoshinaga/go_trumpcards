//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ChinesePokerPresenter チャイニーズポーカープレゼンターインタフェース
type ChinesePokerPresenter interface {
	GamePresenter[interfaces.ChinesePokerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ChinesePokerGame) string
}
