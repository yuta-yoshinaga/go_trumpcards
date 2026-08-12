//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCrazyFourPokerPresenter クレイジー 4 ポーカープレゼンターモック
type MockCrazyFourPokerPresenter struct {
	MockGamePresenter[interfaces.CrazyFourPokerGame]
}

// HintOutput モック
func (_m *MockCrazyFourPokerPresenter) HintOutput(s interfaces.CrazyFourPokerGame) string {
	return _m.Called(s).Get(0).(string)
}
