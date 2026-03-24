package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCribbageInteractor モック
type MockCribbageInteractor struct {
	mock.Mock
}

func (_m *MockCribbageInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCribbageInteractor) ResetWithConfig(cfg domain.CribbageConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockCribbageInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockCribbageInteractor) Peg(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockCribbageInteractor) Go() string {
	return _m.Called().String(0)
}

func (_m *MockCribbageInteractor) ShowNext() string {
	return _m.Called().String(0)
}

func (_m *MockCribbageInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockCribbageInteractor) GetConfig() domain.CribbageConfig {
	return _m.Called().Get(0).(domain.CribbageConfig)
}

func (_m *MockCribbageInteractor) ActionLog() string {
	return _m.Called().String(0)
}
