//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeloteInteractor ベロートインタラクターモック
type MockBeloteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBeloteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBeloteInteractor) ResetWithConfig(cfg domain.BeloteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// PickUp モック
func (_m *MockBeloteInteractor) PickUp(orderUp bool) string {
	ret := _m.Called(orderUp)
	return ret.Get(0).(string)
}

// CallTrump モック
func (_m *MockBeloteInteractor) CallTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockBeloteInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// PassCall モック
func (_m *MockBeloteInteractor) PassCall() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBeloteInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBeloteInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBeloteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBeloteInteractor) GetConfig() domain.BeloteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BeloteConfig)
}

// Hint モック
func (_m *MockBeloteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBeloteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBeloteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
