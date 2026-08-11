//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHokmInteractor ホクムインタラクターモック
type MockHokmInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockHokmInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}

// Reset モック
func (_m *MockHokmInteractor) Reset() string { return _m.Called().Get(0).(string) }

// ResetWithConfig モック
func (_m *MockHokmInteractor) ResetWithConfig(cfg domain.HokmConfig) string {
	return _m.Called(cfg).Get(0).(string)
}

// DeclareTrump モック
func (_m *MockHokmInteractor) DeclareTrump(suit int) string {
	return _m.Called(suit).Get(0).(string)
}

// Play モック
func (_m *MockHokmInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).Get(0).(string)
}

// NextHand モック
func (_m *MockHokmInteractor) NextHand() string { return _m.Called().Get(0).(string) }

// GiveUp モック
func (_m *MockHokmInteractor) GiveUp() string { return _m.Called().Get(0).(string) }

// GetConfig モック
func (_m *MockHokmInteractor) GetConfig() domain.HokmConfig {
	return _m.Called().Get(0).(domain.HokmConfig)
}

// Hint モック
func (_m *MockHokmInteractor) Hint() string { return _m.Called().Get(0).(string) }

// ActionLog モック
func (_m *MockHokmInteractor) ActionLog() string { return _m.Called().Get(0).(string) }
