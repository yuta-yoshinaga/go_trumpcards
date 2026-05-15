//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockCruelPresenter クルーエルプレゼンターモック
type MockCruelPresenter struct {
	mock.Mock
}

func (_m *MockCruelPresenter) Output(c interfaces.CruelGame, lastErr error) string {
	ret := _m.Called(c, lastErr)
	return ret.String(0)
}

func (_m *MockCruelPresenter) HintOutput(c interfaces.CruelGame) string {
	ret := _m.Called(c)
	return ret.String(0)
}

func (_m *MockCruelPresenter) ActionLogOutput(c interfaces.CruelGame) string {
	ret := _m.Called(c)
	return ret.String(0)
}
