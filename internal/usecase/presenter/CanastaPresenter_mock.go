//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCanastaPresenter カナスタプレゼンターモック
type MockCanastaPresenter struct {
	MockGamePresenter[interfaces.CanastaGame]
}

// HintOutput モック
func (_m *MockCanastaPresenter) HintOutput(g interfaces.CanastaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
