//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockEcartePresenter エカルテプレゼンターモック
type MockEcartePresenter struct {
	MockGamePresenter[interfaces.EcarteGame]
}

// HintOutput モック
func (_m *MockEcartePresenter) HintOutput(b interfaces.EcarteGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
