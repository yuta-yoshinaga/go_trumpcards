//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCinchPresenter はチンチプレゼンターモック。
type MockCinchPresenter struct {
	MockGamePresenter[interfaces.CinchGame]
}

// HintOutput モック
func (_m *MockCinchPresenter) HintOutput(g interfaces.CinchGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
