//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockAcesUpPresenter エースアッププレゼンターモック
type MockAcesUpPresenter struct {
	mock.Mock
}

func (_m *MockAcesUpPresenter) Output(g interfaces.AcesUpGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockAcesUpPresenter) HintOutput(g interfaces.AcesUpGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}

func (_m *MockAcesUpPresenter) ActionLogOutput(g interfaces.AcesUpGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
