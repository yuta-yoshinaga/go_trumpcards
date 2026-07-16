//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHighCardFlushPresenter ハイカードフラッシュプレゼンターモック
type MockHighCardFlushPresenter struct {
	MockGamePresenter[interfaces.HighCardFlushGame]
}

// HintOutput モック
func (_m *MockHighCardFlushPresenter) HintOutput(g interfaces.HighCardFlushGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
