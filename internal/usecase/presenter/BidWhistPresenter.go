//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BidWhistPresenter Bid Whist プレゼンターインタフェース
type BidWhistPresenter interface {
	GamePresenter[interfaces.BidWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BidWhistGame) string
}
