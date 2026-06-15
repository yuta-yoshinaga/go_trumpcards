//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSoloWhistInteractor ソロ・ホイストのインタラクターモック
type MockSoloWhistInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSoloWhistInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSoloWhistInteractor) ResetWithConfig(cfg domain.SoloWhistConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockSoloWhistInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSoloWhistInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSoloWhistInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSoloWhistInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSoloWhistInteractor) GetConfig() domain.SoloWhistConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SoloWhistConfig)
}

// Hint モック
func (_m *MockSoloWhistInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSoloWhistInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSoloWhistInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
