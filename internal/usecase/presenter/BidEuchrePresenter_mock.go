//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBidEuchrePresenter ビッド・ユーカー (Bid Euchre) プレゼンターモック
type MockBidEuchrePresenter = MockGamePresenter[interfaces.BidEuchreGame]
