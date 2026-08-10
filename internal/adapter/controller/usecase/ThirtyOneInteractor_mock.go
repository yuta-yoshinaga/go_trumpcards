//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThirtyOneInteractor モック
type MockThirtyOneInteractor struct {
	mock.Mock
}

func (_m *MockThirtyOneInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockThirtyOneInteractor) ResetWithConfig(cfg domain.ThirtyOneConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockThirtyOneInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockThirtyOneInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockThirtyOneInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockThirtyOneInteractor) Knock() string {
	return _m.Called().String(0)
}

func (_m *MockThirtyOneInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockThirtyOneInteractor) GetConfig() domain.ThirtyOneConfig {
	return _m.Called().Get(0).(domain.ThirtyOneConfig)
}

func (_m *MockThirtyOneInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockThirtyOneInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockThirtyOneInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
