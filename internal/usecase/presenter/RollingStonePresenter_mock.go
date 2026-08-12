//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRollingStonePresenter ローリングストーンプレゼンターモック
type MockRollingStonePresenter struct {
	MockGamePresenter[interfaces.RollingStoneGame]
}

// HintOutput モック
func (_m *MockRollingStonePresenter) HintOutput(s interfaces.RollingStoneGame) string {
	return _m.Called(s).Get(0).(string)
}
