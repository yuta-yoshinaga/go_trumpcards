//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRamschPresenter Ramsch presenter mock.
type MockRamschPresenter struct {
	MockGamePresenter[interfaces.RamschGame]
}

// HintOutput mocks the hint output.
func (_m *MockRamschPresenter) HintOutput(s interfaces.RamschGame) string {
	return _m.Called(s).Get(0).(string)
}
