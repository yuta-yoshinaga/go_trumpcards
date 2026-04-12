//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTwoTenJackInteractor ツーテンジャックインタラクターモック
type MockTwoTenJackInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTwoTenJackInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTwoTenJackInteractor) ResetWithConfig(cfg domain.TwoTenJackConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// DeclareTrump モック
func (_m *MockTwoTenJackInteractor) DeclareTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTwoTenJackInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTwoTenJackInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTwoTenJackInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTwoTenJackInteractor) GetConfig() domain.TwoTenJackConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TwoTenJackConfig)
}

// Hint モック
func (_m *MockTwoTenJackInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTwoTenJackInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTwoTenJackInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
