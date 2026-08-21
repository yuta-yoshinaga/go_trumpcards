//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockStHelenaPresenter セント・ヘレナ・ソリティアのプレゼンターモック。
type MockStHelenaPresenter struct {
	MockGamePresenter[interfaces.StHelenaGame]
}

// HintOutput モック。
func (_m *MockStHelenaPresenter) HintOutput(cr interfaces.StHelenaGame) string {
	ret := _m.Called(cr)
	return ret.Get(0).(string)
}
