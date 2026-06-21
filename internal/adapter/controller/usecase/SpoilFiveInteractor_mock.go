//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpoilFiveInteractor スポイル・ファイブのインタラクターモック
type MockSpoilFiveInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpoilFiveInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSpoilFiveInteractor) ResetWithConfig(cfg domain.SpoilFiveConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSpoilFiveInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSpoilFiveInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSpoilFiveInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSpoilFiveInteractor) GetConfig() domain.SpoilFiveConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoilFiveConfig)
}

// Hint モック
func (_m *MockSpoilFiveInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSpoilFiveInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSpoilFiveInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
