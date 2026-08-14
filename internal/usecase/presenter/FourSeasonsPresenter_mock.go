//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockFourSeasonsPresenter フォーシーズンズプレゼンターモック
type MockFourSeasonsPresenter struct {
	mock.Mock
}

func (_m *MockFourSeasonsPresenter) Output(g interfaces.FourSeasonsGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockFourSeasonsPresenter) HintOutput(g interfaces.FourSeasonsGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockFourSeasonsPresenter) ActionLogOutput(g interfaces.FourSeasonsGame) string {
	return _m.Called(g).String(0)
}
