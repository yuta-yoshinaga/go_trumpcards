//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPigPresenter ピッグプレゼンターモック
type MockPigPresenter struct {
	MockGamePresenter[interfaces.PigGame]
}

// HintOutput モック
func (_m *MockPigPresenter) HintOutput(s interfaces.PigGame) string {
	return _m.Called(s).Get(0).(string)
}
