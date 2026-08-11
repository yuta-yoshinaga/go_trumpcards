//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSnapPresenter スナッププレゼンターモック
type MockSnapPresenter struct {
	MockGamePresenter[interfaces.SnapGame]
}

// HintOutput モック
func (_m *MockSnapPresenter) HintOutput(s interfaces.SnapGame) string {
	return _m.Called(s).Get(0).(string)
}
