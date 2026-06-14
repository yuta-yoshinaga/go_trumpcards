//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMusPresenter ムスのプレゼンターモック
type MockMusPresenter struct {
	MockGamePresenter[interfaces.MusGame]
}

// HintOutput モック
func (_m *MockMusPresenter) HintOutput(g interfaces.MusGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
