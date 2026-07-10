//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCalabresellaInteractor カラブレセッラ (Calabresella) のインタラクターモック
type MockCalabresellaInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockCalabresellaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockCalabresellaInteractor) ResetWithConfig(cfg domain.CalabresellaConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

// Bid モック
func (_m *MockCalabresellaInteractor) Bid(bid domain.CalabresellaBid) string {
	ret := _m.Called(bid)
	return ret.Get(0).(string)
}

// Discard モック
func (_m *MockCalabresellaInteractor) Discard(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockCalabresellaInteractor) Play(cardIndex int) string {
	ret := _m.Called(cardIndex)
	return ret.Get(0).(string)
}

// NextTrick モック
func (_m *MockCalabresellaInteractor) NextTrick() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// NextRound モック
func (_m *MockCalabresellaInteractor) NextRound() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// GetConfig モック
func (_m *MockCalabresellaInteractor) GetConfig() domain.CalabresellaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CalabresellaConfig)
}

// Hint モック
func (_m *MockCalabresellaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// ActionLog モック
func (_m *MockCalabresellaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockCalabresellaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
