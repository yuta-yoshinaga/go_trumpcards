//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMacauInteractor モック
type MockMacauInteractor struct {
	mock.Mock
}

func (_m *MockMacauInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockMacauInteractor) ResetWithConfig(cfg domain.MacauConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockMacauInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockMacauInteractor) ChooseSuit(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockMacauInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockMacauInteractor) Declare() string {
	return _m.Called().String(0)
}

func (_m *MockMacauInteractor) SkipDeclare() string {
	return _m.Called().String(0)
}

func (_m *MockMacauInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockMacauInteractor) GetConfig() domain.MacauConfig {
	return _m.Called().Get(0).(domain.MacauConfig)
}

func (_m *MockMacauInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockMacauInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockMacauInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// IsHumanChooseSuitTurn モック
func (_m *MockMacauInteractor) IsHumanChooseSuitTurn() bool {
	return _m.Called().Bool(0)
}

// IsHumanDeclareTurn モック
func (_m *MockMacauInteractor) IsHumanDeclareTurn() bool {
	return _m.Called().Bool(0)
}
