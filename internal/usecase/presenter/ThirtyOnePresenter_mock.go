//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThirtyOnePresenter ThirtyOne プレゼンターモック
type MockThirtyOnePresenter struct {
	MockGamePresenter[interfaces.ThirtyOneGame]
}

// HintOutput モック
func (_m *MockThirtyOnePresenter) HintOutput(g interfaces.ThirtyOneGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
