//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBhabhiPresenter バービープレゼンターモック
type MockBhabhiPresenter struct {
	MockGamePresenter[interfaces.BhabhiGame]
}

// HintOutput モック
func (_m *MockBhabhiPresenter) HintOutput(b interfaces.BhabhiGame) string {
	return _m.Called(b).Get(0).(string)
}
