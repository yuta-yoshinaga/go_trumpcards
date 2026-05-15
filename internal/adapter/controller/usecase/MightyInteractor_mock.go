//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMightyInteractor マイティインタラクターモック
type MockMightyInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockMightyInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockMightyInteractor) ResetWithConfig(cfg domain.MightyConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockMightyInteractor) Bid(bid int, noTrump bool) string {
	ret := _m.Called(bid, noTrump)
	return ret.Get(0).(string)
}

// DeclareTrumpAndFriend モック
func (_m *MockMightyInteractor) DeclareTrumpAndFriend(suit int, partnerSuit int, partnerVal int) string {
	ret := _m.Called(suit, partnerSuit, partnerVal)
	return ret.Get(0).(string)
}

// ExchangeKitty モック
func (_m *MockMightyInteractor) ExchangeKitty(discardIndices []int) string {
	ret := _m.Called(discardIndices)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockMightyInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// PlayJokerLead モック
func (_m *MockMightyInteractor) PlayJokerLead(cardIndex int, demandSuit int) string {
	ret := _m.Called(cardIndex, demandSuit)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockMightyInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockMightyInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockMightyInteractor) GetConfig() domain.MightyConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MightyConfig)
}

// Hint モック
func (_m *MockMightyInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockMightyInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMightyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
