//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBourreInteractor ブーレインタラクターモック
type MockBourreInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockBourreInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockBourreInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Decide モック
func (_m *MockBourreInteractor) Decide(play bool) string {
	ret := _m.Called(play)
	return ret.Get(0).(string)
}

// Draw モック
func (_m *MockBourreInteractor) Draw(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBourreInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextHand モック
func (_m *MockBourreInteractor) NextHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBourreInteractor) ResetWithConfig(config domain.BourreConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBourreInteractor) GetConfig() domain.BourreConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.BourreConfig); ok {
		return val
	}
	return domain.BourreConfig{}
}

// ActionLog モック
func (_m *MockBourreInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
