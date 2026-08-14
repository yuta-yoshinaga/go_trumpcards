//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIronCrossPresenter アイアンクロスプレゼンターモック
type MockIronCrossPresenter struct {
	MockGamePresenter[interfaces.IronCrossGame]
}

// HintOutput モック
func (_m *MockIronCrossPresenter) HintOutput(s interfaces.IronCrossGame) string {
	return _m.Called(s).Get(0).(string)
}
