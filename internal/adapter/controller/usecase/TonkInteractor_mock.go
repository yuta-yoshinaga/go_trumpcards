//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTonkInteractor モック
type MockTonkInteractor struct {
	mock.Mock
}

func (_m *MockTonkInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockTonkInteractor) ResetWithConfig(cfg domain.TonkConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockTonkInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockTonkInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockTonkInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockTonkInteractor) Knock(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockTonkInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockTonkInteractor) GetConfig() domain.TonkConfig {
	return _m.Called().Get(0).(domain.TonkConfig)
}

func (_m *MockTonkInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockTonkInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
