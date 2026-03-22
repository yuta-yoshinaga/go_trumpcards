package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIndianPokerInteractor インディアンポーカーインタラクターモック
type MockIndianPokerInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockIndianPokerInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockIndianPokerInteractor) ResetWithConfig(cfg domain.IndianPokerConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockIndianPokerInteractor) Action(action int, amount int, humanPlayMs int) string {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockIndianPokerInteractor) GetConfig() domain.IndianPokerConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.IndianPokerConfig)
}

// ActionLog モック
func (_m *MockIndianPokerInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
