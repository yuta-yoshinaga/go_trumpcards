//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmiInteractor オミインタラクターモック
type MockOmiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOmiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockOmiInteractor) ResetWithConfig(cfg domain.OmiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// CallTrump モック
func (_m *MockOmiInteractor) CallTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockOmiInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockOmiInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockOmiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockOmiInteractor) GetConfig() domain.OmiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OmiConfig)
}

// Hint モック
func (_m *MockOmiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockOmiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockOmiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
