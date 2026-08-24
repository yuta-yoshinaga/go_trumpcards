//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockQuodlibetInteractor はクオドリベットのインタラクターモック。
type MockQuodlibetInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockQuodlibetInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	b, _ := ret.Get(0).([]byte)
	err, _ := ret.Get(1).(error)
	return b, err
}

// Reset モック
func (_m *MockQuodlibetInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockQuodlibetInteractor) ResetWithConfig(config domain.QuodlibetConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// SelectContract モック
func (_m *MockQuodlibetInteractor) SelectContract(contract int) string {
	ret := _m.Called(contract)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockQuodlibetInteractor) Play(handIdx int) string {
	ret := _m.Called(handIdx)
	return ret.Get(0).(string)
}

// NextDeal モック
func (_m *MockQuodlibetInteractor) NextDeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockQuodlibetInteractor) GetConfig() domain.QuodlibetConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.QuodlibetConfig)
}

// Hint モック
func (_m *MockQuodlibetInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockQuodlibetInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
