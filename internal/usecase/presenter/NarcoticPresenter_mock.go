//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockNarcoticPresenter ナルコティックプレゼンターモック
type MockNarcoticPresenter struct {
	mock.Mock
}

func (_m *MockNarcoticPresenter) Output(g interfaces.NarcoticGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockNarcoticPresenter) HintOutput(g interfaces.NarcoticGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}

func (_m *MockNarcoticPresenter) ActionLogOutput(g interfaces.NarcoticGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
