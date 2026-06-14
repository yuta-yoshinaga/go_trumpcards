//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoppelkopfInteractor ドッペルコップのインタラクターモック
type MockDoppelkopfInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoppelkopfInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDoppelkopfInteractor) ResetWithConfig(cfg domain.DoppelkopfConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDoppelkopfInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Announce モック
func (_m *MockDoppelkopfInteractor) Announce() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockDoppelkopfInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockDoppelkopfInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDoppelkopfInteractor) GetConfig() domain.DoppelkopfConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DoppelkopfConfig)
}

// Hint モック
func (_m *MockDoppelkopfInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDoppelkopfInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockDoppelkopfInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
