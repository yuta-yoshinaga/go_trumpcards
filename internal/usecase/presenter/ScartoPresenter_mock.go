//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockScartoPresenter スカルト (Scarto) のプレゼンターモック
type MockScartoPresenter struct {
	MockGamePresenter[interfaces.ScartoGame]
}

// HintOutput モック
func (_m *MockScartoPresenter) HintOutput(g interfaces.ScartoGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
