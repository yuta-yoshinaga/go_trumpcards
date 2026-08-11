//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockShelemPresenter シェレムプレゼンターモック
type MockShelemPresenter struct {
	MockGamePresenter[interfaces.ShelemGame]
}

// HintOutput モック
func (_m *MockShelemPresenter) HintOutput(s interfaces.ShelemGame) string {
	return _m.Called(s).Get(0).(string)
}
