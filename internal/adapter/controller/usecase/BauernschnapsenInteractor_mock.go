//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBauernschnapsenInteractor バウエルンシュナプセンインタラクターモック
type MockBauernschnapsenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBauernschnapsenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBauernschnapsenInteractor) ResetWithConfig(cfg domain.BauernschnapsenConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// DeclareContract モック
func (_m *MockBauernschnapsenInteractor) DeclareContract(c domain.BauernschnapsenContract, trumpSuit int) string {
	ret := _m.Called(c, trumpSuit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBauernschnapsenInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// DeclareMarriage モック
func (_m *MockBauernschnapsenInteractor) DeclareMarriage(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBauernschnapsenInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBauernschnapsenInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBauernschnapsenInteractor) GetConfig() domain.BauernschnapsenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BauernschnapsenConfig)
}

// Hint モック
func (_m *MockBauernschnapsenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBauernschnapsenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBauernschnapsenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
