//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSedmaPresenter セドマのプレゼンターモック
type MockSedmaPresenter struct {
	MockGamePresenter[interfaces.SedmaGame]
}

// HintOutput モック
func (_m *MockSedmaPresenter) HintOutput(g interfaces.SedmaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
