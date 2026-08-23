//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRistikontraInteractor はリスティコントラ インタラクターモック。
type MockRistikontraInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockRistikontraInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockRistikontraInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockRistikontraInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockRistikontraInteractor) GetConfig() domain.RistikontraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.RistikontraConfig)
}

// ResetWithConfig モック
func (_m *MockRistikontraInteractor) ResetWithConfig(config domain.RistikontraConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockRistikontraInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockRistikontraInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
