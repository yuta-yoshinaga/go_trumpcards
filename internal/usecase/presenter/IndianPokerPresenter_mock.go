//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIndianPokerPresenter インディアンポーカープレゼンターモック
type MockIndianPokerPresenter struct {
	MockGamePresenter[interfaces.IndianPokerGame]
}

// HintOutput モック
func (_m *MockIndianPokerPresenter) HintOutput(g interfaces.IndianPokerGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
