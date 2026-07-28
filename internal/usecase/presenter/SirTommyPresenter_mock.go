//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockSirTommyPresenter サー・トミープレゼンターモック
type MockSirTommyPresenter struct {
	mock.Mock
}

func (_m *MockSirTommyPresenter) Output(g interfaces.SirTommyGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockSirTommyPresenter) HintOutput(g interfaces.SirTommyGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockSirTommyPresenter) ActionLogOutput(g interfaces.SirTommyGame) string {
	return _m.Called(g).String(0)
}
