//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmbreInteractor オンブル (Ombre) のインタラクターモック
type MockOmbreInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockOmbreInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockOmbreInteractor) ResetWithConfig(cfg domain.OmbreConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockOmbreInteractor) Bid(bid domain.OmbreBid, trumpSuit int) string {
	ret := _m.Called(bid, trumpSuit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockOmbreInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockOmbreInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockOmbreInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockOmbreInteractor) GetConfig() domain.OmbreConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbreConfig)
}

// Hint モック
func (_m *MockOmbreInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockOmbreInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockOmbreInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
