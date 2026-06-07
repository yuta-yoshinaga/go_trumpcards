//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFiveHundredPresenter 500プレゼンターモック
type MockFiveHundredPresenter struct {
	MockGamePresenter[interfaces.FiveHundredGame]
}

// HintOutput モック
func (_m *MockFiveHundredPresenter) HintOutput(g interfaces.FiveHundredGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
