//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMinibridgePresenter ミニブリッジプレゼンターモック
type MockMinibridgePresenter struct {
	MockGamePresenter[interfaces.MinibridgeGame]
}

// HintOutput モック
func (_m *MockMinibridgePresenter) HintOutput(s interfaces.MinibridgeGame) string {
	return _m.Called(s).Get(0).(string)
}
