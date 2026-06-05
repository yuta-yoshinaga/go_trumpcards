//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGongZhuInteractor 拱猪インタラクターモック
type MockGongZhuInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockGongZhuInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockGongZhuInteractor) ResetWithConfig(cfg domain.GongZhuConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Expose モック
func (_m *MockGongZhuInteractor) Expose(cardIndices []int) string {
	ret := _m.Called(cardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockGongZhuInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockGongZhuInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockGongZhuInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockGongZhuInteractor) GetConfig() domain.GongZhuConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GongZhuConfig)
}

// Hint モック
func (_m *MockGongZhuInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockGongZhuInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockGongZhuInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
