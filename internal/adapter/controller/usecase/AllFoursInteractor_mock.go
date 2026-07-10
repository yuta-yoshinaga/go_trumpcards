//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAllFoursInteractor All Fours インタラクターモック
type MockAllFoursInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockAllFoursInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockAllFoursInteractor) ResetWithConfig(cfg domain.AllFoursConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Beg モック
func (_m *MockAllFoursInteractor) Beg(beg bool) string {
	ret := _m.Called(beg)
	return ret.Get(0).(string)
}

// RespondBeg モック
func (_m *MockAllFoursInteractor) RespondBeg(run bool) string {
	ret := _m.Called(run)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockAllFoursInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockAllFoursInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockAllFoursInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockAllFoursInteractor) GetConfig() domain.AllFoursConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.AllFoursConfig)
}

// Hint モック
func (_m *MockAllFoursInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockAllFoursInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockAllFoursInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
