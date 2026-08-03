//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMinchiatePresenter ミンキアーテのプレゼンターモック
type MockMinchiatePresenter struct {
	MockGamePresenter[interfaces.MinchiateGame]
}

// HintOutput モック
func (_m *MockMinchiatePresenter) HintOutput(g interfaces.MinchiateGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
