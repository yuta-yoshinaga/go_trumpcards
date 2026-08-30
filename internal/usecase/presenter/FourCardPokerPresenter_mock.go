//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockFourCardPokerPresenter is the mock presenter used in usecase-layer tests.
type MockFourCardPokerPresenter struct {
	MockGamePresenter[interfaces.FourCardPokerGame]
}

// HintOutput mocks the hint output.
func (_m *MockFourCardPokerPresenter) HintOutput(g interfaces.FourCardPokerGame) string {
	ret := _m.Called(g)
	return ret.String(0)
}
