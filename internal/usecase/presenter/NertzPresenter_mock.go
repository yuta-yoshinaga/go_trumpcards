//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockNertzPresenter Nertz プレゼンターモック
type MockNertzPresenter struct {
	mock.Mock
}

func (_m *MockNertzPresenter) Output(g interfaces.NertzGame, lastErr error) string {
	return _m.Called(g, lastErr).String(0)
}

func (_m *MockNertzPresenter) HintOutput(g interfaces.NertzGame) string {
	return _m.Called(g).String(0)
}

func (_m *MockNertzPresenter) ActionLogOutput(g interfaces.NertzGame) string {
	return _m.Called(g).String(0)
}
