//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockToepenPresenter トゥーペン プレゼンターモック
type MockToepenPresenter struct {
	MockGamePresenter[interfaces.ToepenGame]
}

// HintOutput モック
func (_m *MockToepenPresenter) HintOutput(t interfaces.ToepenGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
