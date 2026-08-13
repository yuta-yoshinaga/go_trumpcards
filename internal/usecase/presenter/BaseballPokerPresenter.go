//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BaseballPokerPresenter ベースボールポーカープレゼンターインタフェース
type BaseballPokerPresenter interface {
	GamePresenter[interfaces.BaseballPokerGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.BaseballPokerGame) string
}
