//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEightOffPresenter エイトオフプレゼンターモック
type MockEightOffPresenter struct {
	MockGamePresenter[interfaces.EightOffGame]
}

// HintOutput モック
func (_m *MockEightOffPresenter) HintOutput(e interfaces.EightOffGame) string {
	ret := _m.Called(e)
	return ret.Get(0).(string)
}
