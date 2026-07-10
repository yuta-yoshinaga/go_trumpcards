//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBasraInteractor はバスラ (Basra) のインタラクターモック。
type MockBasraInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockBasraInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockBasraInteractor) ResetWithConfig(cfg domain.BasraConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockBasraInteractor) Play(handIdx int, tableIdxs []int) string {
	ret := _m.Called(handIdx, tableIdxs)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockBasraInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockBasraInteractor) GetConfig() domain.BasraConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BasraConfig)
}

// Hint モック
func (_m *MockBasraInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockBasraInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBasraInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
