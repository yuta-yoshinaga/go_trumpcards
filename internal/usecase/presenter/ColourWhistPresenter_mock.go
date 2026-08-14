//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockColourWhistPresenter カラーホイストプレゼンターモック
type MockColourWhistPresenter struct {
	MockGamePresenter[interfaces.ColourWhistGame]
}

// HintOutput モック
func (_m *MockColourWhistPresenter) HintOutput(s interfaces.ColourWhistGame) string {
	return _m.Called(s).Get(0).(string)
}
