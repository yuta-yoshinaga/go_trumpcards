//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOasisPokerPresenter オアシスポーカープレゼンターモック
type MockOasisPokerPresenter struct {
	MockGamePresenter[interfaces.OasisPokerGame]
}

// HintOutput モック
func (_m *MockOasisPokerPresenter) HintOutput(g interfaces.OasisPokerGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
