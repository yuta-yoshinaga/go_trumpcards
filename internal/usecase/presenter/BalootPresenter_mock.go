//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBalootPresenter バルートプレゼンターモック
type MockBalootPresenter struct {
	MockGamePresenter[interfaces.BalootGame]
}

// HintOutput モック
func (_m *MockBalootPresenter) HintOutput(b interfaces.BalootGame) string {
	return _m.Called(b).Get(0).(string)
}
