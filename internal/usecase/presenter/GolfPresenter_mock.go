//go:build test

package presenter

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MockGolfPresenter ゴルフソリティアプレゼンターモック
type MockGolfPresenter struct {
	mock.Mock
}

func (_m *MockGolfPresenter) Output(g interfaces.GolfGame, lastErr error) string {
	ret := _m.Called(g, lastErr)
	return ret.Get(0).(string)
}

func (_m *MockGolfPresenter) HintOutput(g interfaces.GolfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}

func (_m *MockGolfPresenter) ActionLogOutput(g interfaces.GolfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}

func (_m *MockGolfPresenter) ResetNineHole(g interfaces.GolfGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
