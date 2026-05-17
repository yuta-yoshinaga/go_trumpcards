//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGapsPresenter はGapsPresenterのモック。
type MockGapsPresenter struct {
	MockGamePresenter[interfaces.GapsGame]
}

// HintOutput モック。
func (_m *MockGapsPresenter) HintOutput(g interfaces.GapsGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
