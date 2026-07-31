//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKilleInteractor モック
type MockKilleInteractor struct {
	mock.Mock
}

func (_m *MockKilleInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockKilleInteractor) ResetWithConfig(cfg domain.KilleConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockKilleInteractor) Exchange() string {
	return _m.Called().String(0)
}

func (_m *MockKilleInteractor) Satisfied() string {
	return _m.Called().String(0)
}

func (_m *MockKilleInteractor) Reenter() string {
	return _m.Called().String(0)
}

func (_m *MockKilleInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockKilleInteractor) GetConfig() domain.KilleConfig {
	return _m.Called().Get(0).(domain.KilleConfig)
}

func (_m *MockKilleInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockKilleInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
