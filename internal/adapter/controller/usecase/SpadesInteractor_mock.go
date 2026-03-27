//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpadesInteractor スペードインタラクターモック
type MockSpadesInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpadesInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockSpadesInteractor) ResetWithConfig(cfg domain.SpadesConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockSpadesInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSpadesInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockSpadesInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockSpadesInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockSpadesInteractor) GetConfig() domain.SpadesConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpadesConfig)
}

// Hint モック
func (_m *MockSpadesInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockSpadesInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
