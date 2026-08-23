//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockQuadrilleInteractor カドリール (Quadrille) のインタラクターモック
type MockQuadrilleInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockQuadrilleInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockQuadrilleInteractor) ResetWithConfig(cfg domain.QuadrilleConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockQuadrilleInteractor) Bid(bid domain.QuadrilleBid, trumpSuit int) string {
	ret := _m.Called(bid, trumpSuit)
	return ret.Get(0).(string)
}

// CallKing モック
func (_m *MockQuadrilleInteractor) CallKing(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockQuadrilleInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockQuadrilleInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockQuadrilleInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockQuadrilleInteractor) GetConfig() domain.QuadrilleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleConfig)
}

// Hint モック
func (_m *MockQuadrilleInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockQuadrilleInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockQuadrilleInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
