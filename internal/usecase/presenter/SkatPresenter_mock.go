//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSkatPresenter Skat presenter mock.
type MockSkatPresenter struct {
	MockGamePresenter[interfaces.SkatGame]
}

// HintOutput mocks the hint output.
func (_m *MockSkatPresenter) HintOutput(s interfaces.SkatGame) string {
	return _m.Called(s).Get(0).(string)
}
