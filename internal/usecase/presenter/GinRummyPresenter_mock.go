//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGinRummyPresenter ジンラミープレゼンターモック
type MockGinRummyPresenter struct {
	MockGamePresenter[interfaces.GinRummyGame]
}

// HintOutput モック
func (_m *MockGinRummyPresenter) HintOutput(g interfaces.GinRummyGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
