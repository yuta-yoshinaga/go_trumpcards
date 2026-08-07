//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPaiGowPresenter パイガオポーカープレゼンターモック
type MockPaiGowPresenter struct {
	MockGamePresenter[interfaces.PaiGowGame]
}

// HintOutput モック
func (_m *MockPaiGowPresenter) HintOutput(g interfaces.PaiGowGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
