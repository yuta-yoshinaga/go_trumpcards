//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTysiacInteractor サウザンド (Tysiąc) のインタラクターモック
type MockTysiacInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTysiacInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTysiacInteractor) ResetWithConfig(cfg domain.TysiacConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockTysiacInteractor) Bid(raise bool) string {
	ret := _m.Called(raise)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockTysiacInteractor) Discard(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTysiacInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockTysiacInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockTysiacInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTysiacInteractor) GetConfig() domain.TysiacConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TysiacConfig)
}

// Hint モック
func (_m *MockTysiacInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTysiacInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTysiacInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
