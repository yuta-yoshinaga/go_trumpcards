//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDoubleAttackBlackjackPresenter 追加ベット・ブラックジャックプレゼンターモック
type MockDoubleAttackBlackjackPresenter struct {
	MockGamePresenter[interfaces.DoubleAttackBlackjackGame]
}

// HintOutput モック
func (_m *MockDoubleAttackBlackjackPresenter) HintOutput(s interfaces.DoubleAttackBlackjackGame) string {
	return _m.Called(s).Get(0).(string)
}
