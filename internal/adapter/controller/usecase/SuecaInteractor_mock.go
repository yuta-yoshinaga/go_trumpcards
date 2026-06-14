//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSuecaInteractor スエカのインタラクターモック
type MockSuecaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSuecaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSuecaInteractor) ResetWithConfig(cfg domain.SuecaConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSuecaInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSuecaInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSuecaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSuecaInteractor) GetConfig() domain.SuecaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SuecaConfig)
}

// Hint モック
func (_m *MockSuecaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSuecaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSuecaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
