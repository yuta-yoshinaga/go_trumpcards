//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockChinchonInteractor モック
type MockChinchonInteractor struct {
	mock.Mock
}

func (_m *MockChinchonInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockChinchonInteractor) ResetWithConfig(cfg domain.ChinchonConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockChinchonInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockChinchonInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockChinchonInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockChinchonInteractor) Knock(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockChinchonInteractor) Layoff(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockChinchonInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockChinchonInteractor) GetConfig() domain.ChinchonConfig {
	return _m.Called().Get(0).(domain.ChinchonConfig)
}

func (_m *MockChinchonInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockChinchonInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
