//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWattenInteractor ヴァッテンインタラクターモック
type MockWattenInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockWattenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockWattenInteractor) ResetWithConfig(cfg domain.WattenConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Declare モック
func (_m *MockWattenInteractor) Declare(rank, suit int) string {
	ret := _m.Called(rank, suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockWattenInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Raise モック
func (_m *MockWattenInteractor) Raise() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Respond モック
func (_m *MockWattenInteractor) Respond(hold bool) string {
	ret := _m.Called(hold)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockWattenInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockWattenInteractor) GetConfig() domain.WattenConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.WattenConfig)
}

// Hint モック
func (_m *MockWattenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockWattenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockWattenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
