//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockColoradoPresenter コロラド プレゼンターモック
type MockColoradoPresenter struct {
	MockGamePresenter[interfaces.ColoradoGame]
}

// HintOutput モック
func (_m *MockColoradoPresenter) HintOutput(c interfaces.ColoradoGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
