//go:build test && (!js || !wasm || extra4)

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPutInteractor プットインタラクターモック
type MockPutInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPutInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPutInteractor) ResetWithConfig(cfg domain.PutConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockPutInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Put モック
func (_m *MockPutInteractor) Put() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Respond モック
func (_m *MockPutInteractor) Respond(accept bool) string {
	ret := _m.Called(accept)
	return ret.Get(0).(string)
}

// Next モック
func (_m *MockPutInteractor) Next() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockPutInteractor) GetConfig() domain.PutConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PutConfig)
}

// Hint モック
func (_m *MockPutInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockPutInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPutInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
