//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockVideoPokerPresenter ビデオポーカープレゼンターモック
type MockVideoPokerPresenter struct {
	MockGamePresenter[interfaces.VideoPokerGame]
}

// HintOutput モック
func (_m *MockVideoPokerPresenter) HintOutput(g interfaces.VideoPokerGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
