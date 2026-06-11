//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockCribbagePresenter クリベッジプレゼンターモック
type MockCribbagePresenter struct {
	mock.Mock
}

func (_m *MockCribbagePresenter) Output(g interfaces.CribbageGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockCribbagePresenter) HintOutput(g interfaces.CribbageGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}

func (_m *MockCribbagePresenter) ActionLogOutput(g interfaces.CribbageGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
