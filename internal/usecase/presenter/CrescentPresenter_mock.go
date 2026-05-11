//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCrescentPresenter クレセント・ソリティアのプレゼンターモック。
type MockCrescentPresenter struct {
	MockGamePresenter[interfaces.CrescentGame]
}

// HintOutput モック。
func (_m *MockCrescentPresenter) HintOutput(cr interfaces.CrescentGame) string {
	ret := _m.Called(cr)
	return ret.Get(0).(string)
}
