//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockUltimateTexasHoldemPresenter アルティメット・テキサスホールデムプレゼンターモック
type MockUltimateTexasHoldemPresenter struct {
	MockGamePresenter[interfaces.UltimateTexasHoldemGame]
}

// HintOutput モック
func (_m *MockUltimateTexasHoldemPresenter) HintOutput(g interfaces.UltimateTexasHoldemGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
