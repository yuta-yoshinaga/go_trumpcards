//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTuteInteractor トゥーテのインタラクターモック
type MockTuteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTuteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTuteInteractor) ResetWithConfig(cfg domain.TuteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTuteInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// DeclareMarriage モック
func (_m *MockTuteInteractor) DeclareMarriage(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// DeclareTute モック
func (_m *MockTuteInteractor) DeclareTute() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTuteInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTuteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTuteInteractor) GetConfig() domain.TuteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TuteConfig)
}

// Hint モック
func (_m *MockTuteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTuteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTuteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
