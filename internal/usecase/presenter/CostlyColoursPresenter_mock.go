//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCostlyColoursPresenter はコストリー・カラーズのプレゼンターモック。
type MockCostlyColoursPresenter struct {
	MockGamePresenter[interfaces.CostlyColoursGame]
}

// HintOutput モック
func (_m *MockCostlyColoursPresenter) HintOutput(g interfaces.CostlyColoursGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
