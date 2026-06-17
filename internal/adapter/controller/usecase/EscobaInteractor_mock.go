//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEscobaInteractor エスコバインタラクターモック。
type MockEscobaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockEscobaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockEscobaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockEscobaInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockEscobaInteractor) GetConfig() domain.EscobaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.EscobaConfig)
}

// ResetWithConfig モック
func (_m *MockEscobaInteractor) ResetWithConfig(config domain.EscobaConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockEscobaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockEscobaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
