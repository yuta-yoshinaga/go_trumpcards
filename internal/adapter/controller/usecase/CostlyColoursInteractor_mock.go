//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCostlyColoursInteractor はコストリー・カラーズのインタラクターモック。
type MockCostlyColoursInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockCostlyColoursInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockCostlyColoursInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCostlyColoursInteractor) ResetWithConfig(config domain.CostlyColoursConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Mog モック
func (_m *MockCostlyColoursInteractor) Mog(accept bool) string {
	ret := _m.Called(accept)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCostlyColoursInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// NextDeal モック
func (_m *MockCostlyColoursInteractor) NextDeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCostlyColoursInteractor) GetConfig() domain.CostlyColoursConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CostlyColoursConfig)
}

// Hint モック
func (_m *MockCostlyColoursInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCostlyColoursInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
