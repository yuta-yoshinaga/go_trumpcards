//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FreeBetBlackjackPresenter フリーベット・ブラックジャックプレゼンターインタフェース
type FreeBetBlackjackPresenter interface {
	GamePresenter[interfaces.FreeBetBlackjackGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.FreeBetBlackjackGame) string
}
