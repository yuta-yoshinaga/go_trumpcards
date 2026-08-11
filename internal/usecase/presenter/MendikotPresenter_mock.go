//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMendikotPresenter メンディコットプレゼンターモック
type MockMendikotPresenter struct {
	MockGamePresenter[interfaces.MendikotGame]
}

// HintOutput モック
func (_m *MockMendikotPresenter) HintOutput(m interfaces.MendikotGame) string {
	return _m.Called(m).Get(0).(string)
}
