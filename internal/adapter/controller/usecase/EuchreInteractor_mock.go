package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEuchreInteractor ユーカーインタラクターモック
type MockEuchreInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockEuchreInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockEuchreInteractor) ResetWithConfig(cfg domain.EuchreConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// PickUp モック
func (_m *MockEuchreInteractor) PickUp(orderUp bool, goAlone bool) string {
	ret := _m.Called(orderUp, goAlone)
	return ret.Get(0).(string)
}

// CallTrump モック
func (_m *MockEuchreInteractor) CallTrump(suit int, goAlone bool) string {
	ret := _m.Called(suit, goAlone)
	return ret.Get(0).(string)
}

// PassCall モック
func (_m *MockEuchreInteractor) PassCall() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockEuchreInteractor) Discard(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockEuchreInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockEuchreInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockEuchreInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockEuchreInteractor) GetConfig() domain.EuchreConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.EuchreConfig)
}

// Hint モック
func (_m *MockEuchreInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockEuchreInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
