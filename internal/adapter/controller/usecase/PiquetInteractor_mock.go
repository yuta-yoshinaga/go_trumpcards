//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPiquetInteractor Piquetインタラクターモック
type MockPiquetInteractor struct {
	mock.Mock
}

// Snapshot モック
func (_m *MockPiquetInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	var b []byte
	if v := ret.Get(0); v != nil {
		b = v.([]byte)
	}
	return b, ret.Error(1)
}

// Reset モック
func (_m *MockPiquetInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPiquetInteractor) ResetWithConfig(cfg domain.PiquetConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// ExchangeElder モック
func (_m *MockPiquetInteractor) ExchangeElder(discardIndices []int) string {
	ret := _m.Called(discardIndices)
	return ret.Get(0).(string)
}

// ExchangeYounger モック
func (_m *MockPiquetInteractor) ExchangeYounger(discardIndices []int) string {
	ret := _m.Called(discardIndices)
	return ret.Get(0).(string)
}

// ResolveDeclaration モック
func (_m *MockPiquetInteractor) ResolveDeclaration() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockPiquetInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextDeal モック
func (_m *MockPiquetInteractor) NextDeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Hint モック
func (_m *MockPiquetInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPiquetInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPiquetInteractor) GetConfig() domain.PiquetConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PiquetConfig)
}
