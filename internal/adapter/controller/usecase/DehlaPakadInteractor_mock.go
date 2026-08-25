//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDehlaPakadInteractor はデーラ・パカドのインタラクターモック。
type MockDehlaPakadInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockDehlaPakadInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockDehlaPakadInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDehlaPakadInteractor) ResetWithConfig(config domain.DehlaPakadConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// SelectTrump モック
func (_m *MockDehlaPakadInteractor) SelectTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDehlaPakadInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextHand モック
func (_m *MockDehlaPakadInteractor) NextHand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockDehlaPakadInteractor) GetConfig() domain.DehlaPakadConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DehlaPakadConfig)
}

// Hint モック
func (_m *MockDehlaPakadInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockDehlaPakadInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
