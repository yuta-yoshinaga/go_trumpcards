//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFrenchTarotPresenter フレンチタロット (French Tarot) のプレゼンターモック
type MockFrenchTarotPresenter struct {
	MockGamePresenter[interfaces.FrenchTarotGame]
}

// HintOutput モック
func (_m *MockFrenchTarotPresenter) HintOutput(g interfaces.FrenchTarotGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
