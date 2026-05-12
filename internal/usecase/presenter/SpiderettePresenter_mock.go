//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpiderettePresenter スパイダレットプレゼンターモック
type MockSpiderettePresenter struct {
	MockGamePresenter[interfaces.SpideretteGame]
}

// HintOutput モック
func (_m *MockSpiderettePresenter) HintOutput(s interfaces.SpideretteGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
