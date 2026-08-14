//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBaseballPokerInteractor ベースボールポーカーインタラクターモック
type MockBaseballPokerInteractor struct {
	mock.Mock
}

func (m *MockBaseballPokerInteractor) Reset() string { return m.Called().String(0) }

func (m *MockBaseballPokerInteractor) ResetWithConfig(cfg domain.BaseballPokerConfig) string {
	return m.Called(cfg).String(0)
}

func (m *MockBaseballPokerInteractor) Action(action, amount int) string {
	return m.Called(action, amount).String(0)
}

func (m *MockBaseballPokerInteractor) AnswerBuyIn(answer int) string {
	return m.Called(answer).String(0)
}

func (m *MockBaseballPokerInteractor) NextHand() string { return m.Called().String(0) }

func (m *MockBaseballPokerInteractor) GetConfig() domain.BaseballPokerConfig {
	return m.Called().Get(0).(domain.BaseballPokerConfig)
}

func (m *MockBaseballPokerInteractor) Hint() string { return m.Called().String(0) }

func (m *MockBaseballPokerInteractor) ActionLog() string { return m.Called().String(0) }

// Snapshot モック
func (m *MockBaseballPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).([]byte), ret.Error(1)
}
