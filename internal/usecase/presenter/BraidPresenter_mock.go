//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBraidPresenter ブレイド プレゼンターモック
type MockBraidPresenter struct {
	MockGamePresenter[interfaces.BraidGame]
}

// HintOutput モック
func (_m *MockBraidPresenter) HintOutput(b interfaces.BraidGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
