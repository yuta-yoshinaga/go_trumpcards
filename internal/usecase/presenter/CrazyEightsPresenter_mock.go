//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCrazyEightsPresenter クレイジーエイトプレゼンターモック
type MockCrazyEightsPresenter struct {
	MockGamePresenter[interfaces.CrazyEightsGame]
}

// HintOutput モック
func (_m *MockCrazyEightsPresenter) HintOutput(g interfaces.CrazyEightsGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
