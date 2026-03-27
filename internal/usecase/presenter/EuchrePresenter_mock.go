//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEuchrePresenter ユーカープレゼンターモック
type MockEuchrePresenter struct {
	MockGamePresenter[interfaces.EuchreGame]
}

// HintOutput モック
func (_m *MockEuchrePresenter) HintOutput(e interfaces.EuchreGame) string {
	ret := _m.Called(e)
	return ret.Get(0).(string)
}
