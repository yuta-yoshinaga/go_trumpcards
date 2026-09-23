//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OasisPokerPresenter オアシスポーカープレゼンターインタフェース
type OasisPokerPresenter interface {
	GamePresenter[interfaces.OasisPokerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.OasisPokerGame) string
}
