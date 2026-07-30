//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSkitgubbePresenter シートグッベ プレゼンターモック
type MockSkitgubbePresenter struct {
	MockGamePresenter[interfaces.SkitgubbeGame]
}

// HintOutput モック
func (_m *MockSkitgubbePresenter) HintOutput(c interfaces.SkitgubbeGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
