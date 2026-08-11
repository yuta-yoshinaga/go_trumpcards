//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPasurPresenter パスールプレゼンターモック
type MockPasurPresenter struct {
	MockGamePresenter[interfaces.PasurGame]
}

// HintOutput モック
func (_m *MockPasurPresenter) HintOutput(s interfaces.PasurGame) string {
	return _m.Called(s).Get(0).(string)
}
