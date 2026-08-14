//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRikkenPresenter リッケンプレゼンターモック
type MockRikkenPresenter struct {
	MockGamePresenter[interfaces.RikkenGame]
}

// HintOutput モック
func (_m *MockRikkenPresenter) HintOutput(s interfaces.RikkenGame) string {
	return _m.Called(s).Get(0).(string)
}
