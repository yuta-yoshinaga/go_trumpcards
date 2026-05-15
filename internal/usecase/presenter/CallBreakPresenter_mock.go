//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCallBreakPresenter Call Break プレゼンターモック
type MockCallBreakPresenter struct {
	MockGamePresenter[interfaces.CallBreakGame]
}

// HintOutput モック
func (_m *MockCallBreakPresenter) HintOutput(cb interfaces.CallBreakGame) string {
	ret := _m.Called(cb)
	return ret.Get(0).(string)
}
