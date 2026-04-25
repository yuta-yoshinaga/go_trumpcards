//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockSpiteAndMalicePresenter Spite & Malice プレゼンターモック
type MockSpiteAndMalicePresenter struct {
	mock.Mock
}

func (_m *MockSpiteAndMalicePresenter) Output(g interfaces.SpiteAndMaliceGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockSpiteAndMalicePresenter) HintOutput(g interfaces.SpiteAndMaliceGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockSpiteAndMalicePresenter) ActionLogOutput(g interfaces.SpiteAndMaliceGame) string {
	return _m.Called(g).String(0)
}
