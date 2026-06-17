//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTeenPattiInteractor ティーン・パティのインタラクターモック
type MockTeenPattiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTeenPattiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTeenPattiInteractor) ResetWithConfig(cfg domain.TeenPattiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// See モック
func (_m *MockTeenPattiInteractor) See() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Bet モック
func (_m *MockTeenPattiInteractor) Bet() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Raise モック
func (_m *MockTeenPattiInteractor) Raise(newStake int) string {
	ret := _m.Called(newStake)
	return ret.Get(0).(string)
}

// Fold モック
func (_m *MockTeenPattiInteractor) Fold() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Show モック
func (_m *MockTeenPattiInteractor) Show() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// RequestSideShow モック
func (_m *MockTeenPattiInteractor) RequestSideShow() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// RespondSideShow モック
func (_m *MockTeenPattiInteractor) RespondSideShow(accept bool) string {
	ret := _m.Called(accept)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTeenPattiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTeenPattiInteractor) GetConfig() domain.TeenPattiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TeenPattiConfig)
}

// Hint モック
func (_m *MockTeenPattiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTeenPattiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTeenPattiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
