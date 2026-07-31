//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BidEuchrePresenter ビッド・ユーカー (Bid Euchre) プレゼンターインタフェース
type BidEuchrePresenter = GamePresenter[interfaces.BidEuchreGame]
