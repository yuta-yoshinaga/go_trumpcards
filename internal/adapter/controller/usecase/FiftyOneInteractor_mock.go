//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFiftyOneInteractor フィフティワンインタラクターモック
type MockFiftyOneInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockFiftyOneInteractor) Reset(config domain.FiftyOneConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockFiftyOneInteractor) GetConfig() domain.FiftyOneConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FiftyOneConfig)
}

// ExchangeOne モック
func (_m *MockFiftyOneInteractor) ExchangeOne(handIdx, tableIdx int) string {
	ret := _m.Called(handIdx, tableIdx)
	return ret.Get(0).(string)
}

// ExchangeAll モック
func (_m *MockFiftyOneInteractor) ExchangeAll() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Stop モック
func (_m *MockFiftyOneInteractor) Stop() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockFiftyOneInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFiftyOneInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
