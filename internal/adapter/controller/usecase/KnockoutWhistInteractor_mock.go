//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKnockoutWhistInteractor ノックアウト・ホイストのインタラクターモック
type MockKnockoutWhistInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockKnockoutWhistInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockKnockoutWhistInteractor) ResetWithConfig(cfg domain.KnockoutWhistConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockKnockoutWhistInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockKnockoutWhistInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockKnockoutWhistInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SelectTrump モック
func (_m *MockKnockoutWhistInteractor) SelectTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockKnockoutWhistInteractor) GetConfig() domain.KnockoutWhistConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KnockoutWhistConfig)
}

// Hint モック
func (_m *MockKnockoutWhistInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockKnockoutWhistInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockKnockoutWhistInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
