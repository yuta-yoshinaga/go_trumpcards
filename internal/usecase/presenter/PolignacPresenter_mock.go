//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPolignacPresenter ポリニャックプレゼンターモック
type MockPolignacPresenter struct {
	MockGamePresenter[interfaces.PolignacGame]
}

// HintOutput モック
func (_m *MockPolignacPresenter) HintOutput(p interfaces.PolignacGame) string {
	return _m.Called(p).Get(0).(string)
}
