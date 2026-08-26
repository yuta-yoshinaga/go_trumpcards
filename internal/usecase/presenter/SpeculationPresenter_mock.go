//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpeculationPresenter スペキュレーションプレゼンターモック
type MockSpeculationPresenter struct {
	MockGamePresenter[interfaces.SpeculationGame]
}

// HintOutput モック
func (_m *MockSpeculationPresenter) HintOutput(s interfaces.SpeculationGame) string {
	return _m.Called(s).Get(0).(string)
}
