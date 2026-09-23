//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBinokelPresenter ビノクルプレゼンターモック
type MockBinokelPresenter struct {
	MockGamePresenter[interfaces.BinokelGame]
}

// HintOutput モック
func (_m *MockBinokelPresenter) HintOutput(p interfaces.BinokelGame) string {
	ret := _m.Called(p)
	return ret.Get(0).(string)
}
