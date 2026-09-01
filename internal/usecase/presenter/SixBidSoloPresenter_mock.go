//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSixBidSoloPresenter シックスビッド・ソロ (Six-Bid Solo) プレゼンターモック
type MockSixBidSoloPresenter struct {
	MockGamePresenter[interfaces.SixBidSoloGame]
}

// HintOutput モック
func (_m *MockSixBidSoloPresenter) HintOutput(g interfaces.SixBidSoloGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
