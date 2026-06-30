//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockJassInteractor ヤスインタラクターモック
type MockJassInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockJassInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockJassInteractor) ResetWithConfig(cfg domain.JassConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// ChooseTrump モック
func (_m *MockJassInteractor) ChooseTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Schieben モック
func (_m *MockJassInteractor) Schieben() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockJassInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockJassInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockJassInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockJassInteractor) GetConfig() domain.JassConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.JassConfig)
}

// Hint モック
func (_m *MockJassInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockJassInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockJassInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
