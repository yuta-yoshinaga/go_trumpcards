//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrazyFourPokerInteractor クレイジー 4 ポーカーインタラクターモック
type MockCrazyFourPokerInteractor struct {
	mock.Mock
}

func (m *MockCrazyFourPokerInteractor) Reset() string { return m.Called().String(0) }

func (m *MockCrazyFourPokerInteractor) ResetWithConfig(cfg domain.CrazyFourPokerConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockCrazyFourPokerInteractor) PlaceBet(ante, queensUp int) string {
	return m.Called(ante, queensUp).String(0)
}

func (m *MockCrazyFourPokerInteractor) Play(multiplier int) string {
	return m.Called(multiplier).String(0)
}

func (m *MockCrazyFourPokerInteractor) Fold() string { return m.Called().String(0) }

func (m *MockCrazyFourPokerInteractor) NextRound() string { return m.Called().String(0) }

func (m *MockCrazyFourPokerInteractor) GetConfig() domain.CrazyFourPokerConfig {
	return m.Called().Get(0).(domain.CrazyFourPokerConfig)
}

func (m *MockCrazyFourPokerInteractor) Hint() string { return m.Called().String(0) }

func (m *MockCrazyFourPokerInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockCrazyFourPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
