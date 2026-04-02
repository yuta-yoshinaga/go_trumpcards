//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPinochlePresenter ピノクルプレゼンターモック
type MockPinochlePresenter struct {
	MockGamePresenter[interfaces.PinochleGame]
}

// HintOutput モック
func (_m *MockPinochlePresenter) HintOutput(p interfaces.PinochleGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
