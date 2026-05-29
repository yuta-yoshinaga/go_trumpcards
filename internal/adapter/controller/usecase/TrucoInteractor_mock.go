//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrucoInteractor トゥルコインタラクターモック
type MockTrucoInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockTrucoInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockTrucoInteractor) ResetWithConfig(cfg domain.TrucoConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockTrucoInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Truco モック
func (_m *MockTrucoInteractor) Truco() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Respond モック
func (_m *MockTrucoInteractor) Respond(accept bool) string {
	ret := _m.Called(accept)
	return ret.Get(0).(string)
}

// Next モック
func (_m *MockTrucoInteractor) Next() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockTrucoInteractor) GetConfig() domain.TrucoConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TrucoConfig)
}

// Hint モック
func (_m *MockTrucoInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockTrucoInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTrucoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
