//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSchnapsenInteractor シュナプセンインタラクターモック
type MockSchnapsenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSchnapsenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSchnapsenInteractor) ResetWithConfig(cfg domain.SchnapsenConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSchnapsenInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// DeclareMarriage モック
func (_m *MockSchnapsenInteractor) DeclareMarriage(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSchnapsenInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSchnapsenInteractor) GetConfig() domain.SchnapsenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SchnapsenConfig)
}

// Hint モック
func (_m *MockSchnapsenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSchnapsenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSchnapsenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
