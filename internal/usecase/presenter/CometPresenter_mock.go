//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCometPresenter はコメットのプレゼンターモック。
type MockCometPresenter struct {
	MockGamePresenter[interfaces.CometGame]
}

// HintOutput モック
func (_m *MockCometPresenter) HintOutput(g interfaces.CometGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
