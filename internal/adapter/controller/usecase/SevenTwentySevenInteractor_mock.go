//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenTwentySevenInteractor はセブン・トゥエンティセブン (SevenTwentySeven) のインタラクターモック。
type MockSevenTwentySevenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSevenTwentySevenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSevenTwentySevenInteractor) ResetWithConfig(cfg domain.SevenTwentySevenConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// TakeCard モック
func (_m *MockSevenTwentySevenInteractor) TakeCard(draw bool) string {
	return _m.Called(draw).Get(0).(string)
}

// NextRound モック
func (_m *MockSevenTwentySevenInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSevenTwentySevenInteractor) GetConfig() domain.SevenTwentySevenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SevenTwentySevenConfig)
}

// Hint モック
func (_m *MockSevenTwentySevenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSevenTwentySevenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSevenTwentySevenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
