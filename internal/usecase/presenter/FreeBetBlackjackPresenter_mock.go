//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFreeBetBlackjackPresenter フリーベット・ブラックジャックプレゼンターモック
type MockFreeBetBlackjackPresenter struct {
	MockGamePresenter[interfaces.FreeBetBlackjackGame]
}

// HintOutput モック
func (_m *MockFreeBetBlackjackPresenter) HintOutput(s interfaces.FreeBetBlackjackGame) string {
	return _m.Called(s).Get(0).(string)
}
