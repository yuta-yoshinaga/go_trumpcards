//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHorsePresenter は H.O.R.S.E. のプレゼンターモック。
type MockHorsePresenter struct {
	MockGamePresenter[interfaces.HorseGame]
}

// HintOutput モック
func (_m *MockHorsePresenter) HintOutput(g interfaces.HorseGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
