//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// VideoPokerPresenter ビデオポーカープレゼンターインタフェース
type VideoPokerPresenter interface {
	GamePresenter[interfaces.VideoPokerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.VideoPokerGame) string
}
