//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSchafkopfInteractor シャーフコップのインタラクターモック
type MockSchafkopfInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSchafkopfInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSchafkopfInteractor) ResetWithConfig(cfg domain.SchafkopfConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Declare モック
func (_m *MockSchafkopfInteractor) Declare(pick bool, contract domain.SchafkopfContract, soloSuit int) string {
	ret := _m.Called(pick, contract, soloSuit)
	return ret.Get(0).(string)
}

// Call モック
func (_m *MockSchafkopfInteractor) Call(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSchafkopfInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSchafkopfInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSchafkopfInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSchafkopfInteractor) GetConfig() domain.SchafkopfConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SchafkopfConfig)
}

// Hint モック
func (_m *MockSchafkopfInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSchafkopfInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSchafkopfInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
