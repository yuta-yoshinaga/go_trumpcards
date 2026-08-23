//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMadrassoPresenter マドラッソのプレゼンターモック
type MockMadrassoPresenter struct {
	MockGamePresenter[interfaces.MadrassoGame]
}

// HintOutput モック
func (_m *MockMadrassoPresenter) HintOutput(g interfaces.MadrassoGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
