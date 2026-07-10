//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAllFoursPresenter All Fours プレゼンターモック
type MockAllFoursPresenter struct {
	MockGamePresenter[interfaces.AllFoursGame]
}

// HintOutput モック
func (_m *MockAllFoursPresenter) HintOutput(g interfaces.AllFoursGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
