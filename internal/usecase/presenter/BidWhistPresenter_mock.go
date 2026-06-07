//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBidWhistPresenter Bid Whist プレゼンターモック
type MockBidWhistPresenter struct {
	MockGamePresenter[interfaces.BidWhistGame]
}

// HintOutput モック
func (_m *MockBidWhistPresenter) HintOutput(g interfaces.BidWhistGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
