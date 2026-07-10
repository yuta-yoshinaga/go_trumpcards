//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKingInteractor はキングインタラクターモック。
type MockKingInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockKingInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextDeal モック
func (_m *MockKingInteractor) NextDeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SelectContract モック
func (_m *MockKingInteractor) SelectContract(contract, trumpSuit int) string {
	ret := _m.Called(contract, trumpSuit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockKingInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockKingInteractor) GetConfig() domain.KingConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KingConfig)
}

// ResetWithConfig モック
func (_m *MockKingInteractor) ResetWithConfig(config domain.KingConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockKingInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockKingInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockKingInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
