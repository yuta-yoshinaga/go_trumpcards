package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGinRummyInteractor モック
type MockGinRummyInteractor struct {
	mock.Mock
}

func (_m *MockGinRummyInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockGinRummyInteractor) ResetWithConfig(cfg domain.GinRummyConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockGinRummyInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockGinRummyInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockGinRummyInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockGinRummyInteractor) Knock(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockGinRummyInteractor) Layoff(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockGinRummyInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockGinRummyInteractor) GetConfig() domain.GinRummyConfig {
	return _m.Called().Get(0).(domain.GinRummyConfig)
}

func (_m *MockGinRummyInteractor) ActionLog() string {
	return _m.Called().String(0)
}
