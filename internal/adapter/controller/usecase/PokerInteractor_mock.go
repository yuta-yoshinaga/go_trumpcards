package usecase

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPokerInteractor ポーカーインタラクターモック
type MockPokerInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockPokerInteractor) ResetWithConfig(cfg domain.PokerConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Action モック
func (_m *MockPokerInteractor) Action(action int, amount int) string {
	ret := _m.Called(action, amount)
	return ret.Get(0).(string)
}

// Exchange モック
func (_m *MockPokerInteractor) Exchange(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// Stand モック
func (_m *MockPokerInteractor) Stand() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
