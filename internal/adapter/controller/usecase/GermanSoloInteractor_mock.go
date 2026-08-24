//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGermanSoloInteractor ジャーマン・ソロ (GermanSolo) のインタラクターモック
type MockGermanSoloInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGermanSoloInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGermanSoloInteractor) ResetWithConfig(cfg domain.GermanSoloConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockGermanSoloInteractor) Bid(bid domain.GermanSoloBid, trumpSuit int) string {
	ret := _m.Called(bid, trumpSuit)
	return ret.Get(0).(string)
}

// CallAce モック
func (_m *MockGermanSoloInteractor) CallAce(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGermanSoloInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockGermanSoloInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGermanSoloInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGermanSoloInteractor) GetConfig() domain.GermanSoloConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloConfig)
}

// Hint モック
func (_m *MockGermanSoloInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGermanSoloInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGermanSoloInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
