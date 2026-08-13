//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBaseballPokerPresenter ベースボールポーカープレゼンターモック
type MockBaseballPokerPresenter struct {
	MockGamePresenter[interfaces.BaseballPokerGame]
}

// HintOutput モック
func (_m *MockBaseballPokerPresenter) HintOutput(s interfaces.BaseballPokerGame) string {
	return _m.Called(s).Get(0).(string)
}
