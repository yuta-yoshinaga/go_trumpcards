//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNinetyNineInteractor ナインティナインインタラクターモック
type MockNinetyNineInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockNinetyNineInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockNinetyNineInteractor) ResetWithConfig(cfg domain.NinetyNineConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockNinetyNineInteractor) Bid(buryIndices []int) string {
	ret := _m.Called(buryIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockNinetyNineInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockNinetyNineInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockNinetyNineInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockNinetyNineInteractor) GetConfig() domain.NinetyNineConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.NinetyNineConfig)
}

// Hint モック
func (_m *MockNinetyNineInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockNinetyNineInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockNinetyNineInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
