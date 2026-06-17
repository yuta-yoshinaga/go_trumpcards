//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEcarteInteractor エカルテインタラクターモック
type MockEcarteInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockEcarteInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockEcarteInteractor) ResetWithConfig(cfg domain.EcarteConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Propose モック
func (_m *MockEcarteInteractor) Propose() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockEcarteInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Respond モック
func (_m *MockEcarteInteractor) Respond(accept bool) string {
	ret := _m.Called(accept)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockEcarteInteractor) Discard(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockEcarteInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockEcarteInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockEcarteInteractor) GetConfig() domain.EcarteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.EcarteConfig)
}

// Hint モック
func (_m *MockEcarteInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockEcarteInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockEcarteInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
