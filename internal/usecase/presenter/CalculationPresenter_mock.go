//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockCalculationPresenter カルキュレーションプレゼンターモック
type MockCalculationPresenter struct {
	mock.Mock
}

func (_m *MockCalculationPresenter) Output(g interfaces.CalculationGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockCalculationPresenter) HintOutput(g interfaces.CalculationGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockCalculationPresenter) ActionLogOutput(g interfaces.CalculationGame) string {
	return _m.Called(g).String(0)
}
