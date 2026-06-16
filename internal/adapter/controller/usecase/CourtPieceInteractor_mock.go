//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCourtPieceInteractor Court Piece インタラクターモック
type MockCourtPieceInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCourtPieceInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCourtPieceInteractor) ResetWithConfig(cfg domain.CourtPieceConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// DeclareTrump モック
func (_m *MockCourtPieceInteractor) DeclareTrump(suit int) string {
	ret := _m.Called(suit)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCourtPieceInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockCourtPieceInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCourtPieceInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCourtPieceInteractor) GetConfig() domain.CourtPieceConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CourtPieceConfig)
}

// Hint モック
func (_m *MockCourtPieceInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCourtPieceInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCourtPieceInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
