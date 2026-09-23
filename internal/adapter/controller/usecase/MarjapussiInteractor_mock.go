//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMarjapussiInteractor マルヤプッシ (Marjapussi) のインタラクターモック
type MockMarjapussiInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMarjapussiInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMarjapussiInteractor) ResetWithConfig(cfg domain.MarjapussiConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockMarjapussiInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockMarjapussiInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockMarjapussiInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMarjapussiInteractor) GetConfig() domain.MarjapussiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MarjapussiConfig)
}

// Hint モック
func (_m *MockMarjapussiInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMarjapussiInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMarjapussiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
