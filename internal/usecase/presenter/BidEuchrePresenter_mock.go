//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBidEuchrePresenter ビッド・ユーカー (Bid Euchre) プレゼンターモック
type MockBidEuchrePresenter struct {
	MockGamePresenter[interfaces.BidEuchreGame]
}

// HintOutput モック
func (_m *MockBidEuchrePresenter) HintOutput(g interfaces.BidEuchreGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
