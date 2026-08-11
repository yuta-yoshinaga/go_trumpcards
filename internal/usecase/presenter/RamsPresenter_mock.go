//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRamsPresenter ラムスプレゼンターモック
type MockRamsPresenter struct {
	MockGamePresenter[interfaces.RamsGame]
}

// HintOutput モック
func (_m *MockRamsPresenter) HintOutput(r interfaces.RamsGame) string {
	return _m.Called(r).Get(0).(string)
}
