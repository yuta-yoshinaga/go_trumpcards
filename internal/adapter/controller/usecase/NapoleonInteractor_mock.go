//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNapoleonInteractor ナポレオンインタラクターモック
type MockNapoleonInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockNapoleonInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockNapoleonInteractor) ResetWithConfig(cfg domain.NapoleonConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockNapoleonInteractor) Bid(bid int) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// DeclareTrump モック
func (_m *MockNapoleonInteractor) DeclareTrump(suit int, adjSuit int, adjVal int) string {
	ret := _m.Called(suit, adjSuit, adjVal)
	return ret.Get(0).(string)
}

// ExchangeKitty モック
func (_m *MockNapoleonInteractor) ExchangeKitty(discardIndex int) string {
	ret := _m.Called(discardIndex)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockNapoleonInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockNapoleonInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockNapoleonInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockNapoleonInteractor) GetConfig() domain.NapoleonConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.NapoleonConfig)
}

// Hint モック
func (_m *MockNapoleonInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockNapoleonInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}
