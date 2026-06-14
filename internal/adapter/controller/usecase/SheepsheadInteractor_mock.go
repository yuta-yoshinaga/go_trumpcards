//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSheepsheadInteractor シープスヘッドのインタラクターモック
type MockSheepsheadInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSheepsheadInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSheepsheadInteractor) ResetWithConfig(cfg domain.SheepsheadConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Pick モック
func (_m *MockSheepsheadInteractor) Pick(pick bool) string {
	ret := _m.Called(pick)
	return ret.Get(0).(string)
}

// Bury モック
func (_m *MockSheepsheadInteractor) Bury(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Call モック
func (_m *MockSheepsheadInteractor) Call(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSheepsheadInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSheepsheadInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSheepsheadInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSheepsheadInteractor) GetConfig() domain.SheepsheadConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SheepsheadConfig)
}

// Hint モック
func (_m *MockSheepsheadInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSheepsheadInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSheepsheadInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
