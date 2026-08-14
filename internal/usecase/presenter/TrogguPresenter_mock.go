//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTrogguPresenter トロッグのプレゼンターモック。
type MockTrogguPresenter struct {
	MockGamePresenter[interfaces.TrogguGame]
}

// HintOutput モック
func (_m *MockTrogguPresenter) HintOutput(g interfaces.TrogguGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
