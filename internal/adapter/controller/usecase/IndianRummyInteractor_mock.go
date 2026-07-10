//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIndianRummyInteractor モック
type MockIndianRummyInteractor struct {
	mock.Mock
}

func (_m *MockIndianRummyInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockIndianRummyInteractor) ResetWithConfig(cfg domain.IndianRummyConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockIndianRummyInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockIndianRummyInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockIndianRummyInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockIndianRummyInteractor) Declare(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockIndianRummyInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockIndianRummyInteractor) GetConfig() domain.IndianRummyConfig {
	return _m.Called().Get(0).(domain.IndianRummyConfig)
}

func (_m *MockIndianRummyInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockIndianRummyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
