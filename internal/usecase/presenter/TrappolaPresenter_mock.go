//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTrappolaPresenter トラッポラのプレゼンターモック
type MockTrappolaPresenter struct {
	MockGamePresenter[interfaces.TrappolaGame]
}

// HintOutput モック
func (_m *MockTrappolaPresenter) HintOutput(g interfaces.TrappolaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
