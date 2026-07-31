//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSixBidSoloPresenter シックスビッド・ソロ (Six-Bid Solo) プレゼンターモック
type MockSixBidSoloPresenter = MockGamePresenter[interfaces.SixBidSoloGame]
