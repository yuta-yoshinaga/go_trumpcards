//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCrazyQuiltPresenter クレイジーキルト プレゼンターモック
type MockCrazyQuiltPresenter struct {
	MockGamePresenter[interfaces.CrazyQuiltGame]
}

// HintOutput モック
func (_m *MockCrazyQuiltPresenter) HintOutput(c interfaces.CrazyQuiltGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
