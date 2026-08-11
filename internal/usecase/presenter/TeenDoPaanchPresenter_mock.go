//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTeenDoPaanchPresenter 3-2-5 プレゼンターモック
type MockTeenDoPaanchPresenter struct {
	MockGamePresenter[interfaces.TeenDoPaanchGame]
}

// HintOutput モック
func (_m *MockTeenDoPaanchPresenter) HintOutput(g interfaces.TeenDoPaanchGame) string {
	return _m.Called(g).Get(0).(string)
}
