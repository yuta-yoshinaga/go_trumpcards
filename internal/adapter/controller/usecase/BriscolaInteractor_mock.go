//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBriscolaInteractor ブリスコラインタラクターモック
type MockBriscolaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBriscolaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBriscolaInteractor) ResetWithConfig(cfg domain.BriscolaConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBriscolaInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockBriscolaInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBriscolaInteractor) GetConfig() domain.BriscolaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BriscolaConfig)
}

// Hint モック
func (_m *MockBriscolaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBriscolaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBriscolaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
