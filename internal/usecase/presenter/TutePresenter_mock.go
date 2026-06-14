//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTutePresenter トゥーテのプレゼンターモック
type MockTutePresenter struct {
	MockGamePresenter[interfaces.TuteGame]
}

// HintOutput モック
func (_m *MockTutePresenter) HintOutput(g interfaces.TuteGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
