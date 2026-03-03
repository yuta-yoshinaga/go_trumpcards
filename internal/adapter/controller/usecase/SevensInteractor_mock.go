package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevensInteractor 7並べインタラクターモック
type MockSevensInteractor struct {
	mock.Mock
}

// ResetWithConfig モック
func (_m *MockSevensInteractor) ResetWithConfig(cfg domain.SevensConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Reset モック
func (_m *MockSevensInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSevensInteractor) Play(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

// PlayJoker モック
func (_m *MockSevensInteractor) PlayJoker(cardIdx, targetSuit, targetValue int) string {
	ret := _m.Called(cardIdx, targetSuit, targetValue)
	return ret.Get(0).(string)
}
