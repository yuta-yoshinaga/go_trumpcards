//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BidEuchrePresenter ビッド・ユーカー (Bid Euchre) プレゼンターインタフェース
//
// **ヒントは CUI 専用** (#5730)。Web 側は既存の状態出力をそのまま返す。
type BidEuchrePresenter interface {
	GamePresenter[interfaces.BidEuchreGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BidEuchreGame) string
}
