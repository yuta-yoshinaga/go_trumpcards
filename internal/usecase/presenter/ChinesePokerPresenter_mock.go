//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockChinesePokerPresenter チャイニーズポーカープレゼンターモック
type MockChinesePokerPresenter struct {
	MockGamePresenter[interfaces.ChinesePokerGame]
}

// HintOutput モック
func (_m *MockChinesePokerPresenter) HintOutput(g interfaces.ChinesePokerGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
