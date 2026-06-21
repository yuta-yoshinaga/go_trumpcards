//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeziqueInteractor ベジークインタラクターモック
type MockBeziqueInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBeziqueInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBeziqueInteractor) ResetWithConfig(cfg domain.BeziqueConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBeziqueInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// DeclareMeld モック
func (_m *MockBeziqueInteractor) DeclareMeld(meldIndex int) string {
	ret := _m.Called(meldIndex)
	return ret.Get(0).(string)
}

// SkipMeld モック
func (_m *MockBeziqueInteractor) SkipMeld() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBeziqueInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBeziqueInteractor) GetConfig() domain.BeziqueConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BeziqueConfig)
}

// Hint モック
func (_m *MockBeziqueInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBeziqueInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBeziqueInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
