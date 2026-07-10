//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGoStopPresenter はゴーストッププレゼンターモック。
type MockGoStopPresenter struct {
	MockGamePresenter[interfaces.GoStopGame]
}

// HintOutput モック
func (_m *MockGoStopPresenter) HintOutput(g interfaces.GoStopGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
