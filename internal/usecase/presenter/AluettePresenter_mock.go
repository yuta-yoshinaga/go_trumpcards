//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockAluettePresenter アリュエットのプレゼンターモック
type MockAluettePresenter struct {
	MockGamePresenter[interfaces.AluetteGame]
}

// HintOutput モック
func (_m *MockAluettePresenter) HintOutput(g interfaces.AluetteGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
