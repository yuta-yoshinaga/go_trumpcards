//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrazyEightsInteractor モック
type MockCrazyEightsInteractor struct {
	mock.Mock
}

func (_m *MockCrazyEightsInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCrazyEightsInteractor) ResetWithConfig(cfg domain.CrazyEightsConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockCrazyEightsInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockCrazyEightsInteractor) ChooseSuit(suit int) string {
	return _m.Called(suit).String(0)
}

func (_m *MockCrazyEightsInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockCrazyEightsInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockCrazyEightsInteractor) GetConfig() domain.CrazyEightsConfig {
	return _m.Called().Get(0).(domain.CrazyEightsConfig)
}

func (_m *MockCrazyEightsInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockCrazyEightsInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockCrazyEightsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// IsHumanChooseSuitTurn モック
func (_m *MockCrazyEightsInteractor) IsHumanChooseSuitTurn() bool {
	return _m.Called().Bool(0)
}
