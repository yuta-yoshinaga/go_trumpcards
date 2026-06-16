//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBeziquePresenter ベジークプレゼンターモック
type MockBeziquePresenter struct {
	MockGamePresenter[interfaces.BeziqueGame]
}

// HintOutput モック
func (_m *MockBeziquePresenter) HintOutput(b interfaces.BeziqueGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
