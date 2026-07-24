//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDeuceToSevenPresenter is the testify/mock presenter for 2-7 Triple Draw.
type MockDeuceToSevenPresenter struct {
	MockGamePresenter[interfaces.DeuceToSevenGame]
}

// HintOutput mock.
func (_m *MockDeuceToSevenPresenter) HintOutput(g interfaces.DeuceToSevenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
