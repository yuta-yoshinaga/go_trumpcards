//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSevenCardStudPresenter セブンカードスタッドプレゼンターモック
type MockSevenCardStudPresenter struct {
	MockGamePresenter[interfaces.SevenCardStudGame]
}

// HintOutput モック
func (_m *MockSevenCardStudPresenter) HintOutput(g interfaces.SevenCardStudGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
