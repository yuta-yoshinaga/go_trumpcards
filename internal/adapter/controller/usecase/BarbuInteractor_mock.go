//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBarbuInteractor はバルブインタラクターモック。
type MockBarbuInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBarbuInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextDeal モック
func (_m *MockBarbuInteractor) NextDeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// SelectContract モック
func (_m *MockBarbuInteractor) SelectContract(contract, trumpSuit int) string {
	ret := _m.Called(contract, trumpSuit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBarbuInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBarbuInteractor) GetConfig() domain.BarbuConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BarbuConfig)
}

// ResetWithConfig モック
func (_m *MockBarbuInteractor) ResetWithConfig(config domain.BarbuConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBarbuInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBarbuInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
