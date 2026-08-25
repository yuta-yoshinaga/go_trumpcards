//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockContinentalRummyInteractor はコンチネンタル・ラミーのインタラクターモック。
type MockContinentalRummyInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockContinentalRummyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockContinentalRummyInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockContinentalRummyInteractor) ResetWithConfig(config domain.ContinentalRummyConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// DrawStock モック
func (_m *MockContinentalRummyInteractor) DrawStock() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// DrawDiscard モック
func (_m *MockContinentalRummyInteractor) DrawDiscard() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockContinentalRummyInteractor) Discard(i int) string {
	ret := _m.Called(i)
	return ret.Get(0).(string)
}

// GoOut モック
func (_m *MockContinentalRummyInteractor) GoOut(i int) string {
	ret := _m.Called(i)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockContinentalRummyInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockContinentalRummyInteractor) GetConfig() domain.ContinentalRummyConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ContinentalRummyConfig)
}

// Hint モック
func (_m *MockContinentalRummyInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockContinentalRummyInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
