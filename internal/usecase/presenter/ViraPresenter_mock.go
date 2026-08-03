//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockViraPresenter ヴィーラのプレゼンターモック
type MockViraPresenter struct {
	MockGamePresenter[interfaces.ViraGame]
}

// HintOutput モック
func (_m *MockViraPresenter) HintOutput(g interfaces.ViraGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
