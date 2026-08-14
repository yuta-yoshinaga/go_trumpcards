//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockZwanzigerrufenInteractor ツヴァンツィガールーフェンのインタラクターモック。
type MockZwanzigerrufenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockZwanzigerrufenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockZwanzigerrufenInteractor) ResetWithConfig(cfg domain.ZwanzigerrufenConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockZwanzigerrufenInteractor) Bid(bid domain.ZwanzigerrufenBid) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Pass モック
func (_m *MockZwanzigerrufenInteractor) Pass() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockZwanzigerrufenInteractor) Discard(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockZwanzigerrufenInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockZwanzigerrufenInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockZwanzigerrufenInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockZwanzigerrufenInteractor) GetConfig() domain.ZwanzigerrufenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ZwanzigerrufenConfig)
}

// Hint モック
func (_m *MockZwanzigerrufenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockZwanzigerrufenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockZwanzigerrufenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
