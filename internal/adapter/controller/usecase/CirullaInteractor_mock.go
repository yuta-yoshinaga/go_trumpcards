//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCirullaInteractor はチルッラのインタラクターモック。
type MockCirullaInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockCirullaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockCirullaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCirullaInteractor) ResetWithConfig(config domain.CirullaConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCirullaInteractor) Play(handIdx int, captureIdxs []int) string {
	ret := _m.Called(handIdx, captureIdxs)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCirullaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCirullaInteractor) GetConfig() domain.CirullaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CirullaConfig)
}

// Hint モック
func (_m *MockCirullaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCirullaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
