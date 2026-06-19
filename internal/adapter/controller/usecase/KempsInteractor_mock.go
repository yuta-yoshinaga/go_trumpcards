//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKempsInteractor はケムプスのインタラクターモック。
type MockKempsInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockKempsInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockKempsInteractor) ResetWithConfig(cfg domain.KempsConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Swap モック
func (_m *MockKempsInteractor) Swap(handIndex, fieldIndex int) string {
	ret := _m.Called(handIndex, fieldIndex)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockKempsInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SetSignal モック
func (_m *MockKempsInteractor) SetSignal(signalType int) string {
	ret := _m.Called(signalType)
	return ret.Get(0).(string)
}

// DeclareKemps モック
func (_m *MockKempsInteractor) DeclareKemps() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DeclareCounterKemps モック
func (_m *MockKempsInteractor) DeclareCounterKemps(targetSeat int) string {
	ret := _m.Called(targetSeat)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockKempsInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockKempsInteractor) GetConfig() domain.KempsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KempsConfig)
}

// ActionLog モック
func (_m *MockKempsInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockKempsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
