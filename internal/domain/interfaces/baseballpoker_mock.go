//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBaseballPokerGame ベースボールポーカーゲームモック
type MockBaseballPokerGame struct {
	mock.Mock
}

func (m *MockBaseballPokerGame) Reset() { m.Called() }

func (m *MockBaseballPokerGame) PlayerAction(action, amount int) error {
	return m.Called(action, amount).Error(0)
}

func (m *MockBaseballPokerGame) AnswerBuyIn(answer int) error {
	return m.Called(answer).Error(0)
}

func (m *MockBaseballPokerGame) NextHand() error { return m.Called().Error(0) }

func (m *MockBaseballPokerGame) CpuPlay() { m.Called() }

func (m *MockBaseballPokerGame) GetConfig() domain.BaseballPokerConfig {
	return m.Called().Get(0).(domain.BaseballPokerConfig)
}

func (m *MockBaseballPokerGame) SetConfig(cfg domain.BaseballPokerConfig) { m.Called(cfg) }

func (m *MockBaseballPokerGame) GetPhase() domain.BaseballPhase {
	return m.Called().Get(0).(domain.BaseballPhase)
}

func (m *MockBaseballPokerGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockBaseballPokerGame) GetPlayers() []*domain.BaseballPokerPlayer {
	return m.Called().Get(0).([]*domain.BaseballPokerPlayer)
}

func (m *MockBaseballPokerGame) GetStreet() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetPot() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetCurrentBet() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetToCall() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetRaiseCount() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) CanRaise() bool { return m.Called().Bool(0) }

func (m *MockBaseballPokerGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetBuyerSeat() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetBuyCost() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) HumanSeat() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockBaseballPokerGame) IsHumanBuying() bool { return m.Called().Bool(0) }

func (m *MockBaseballPokerGame) GetHandNumber() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetResults() []domain.BaseballResult {
	return m.Called().Get(0).([]domain.BaseballResult)
}

func (m *MockBaseballPokerGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockBaseballPokerGame) GetHint() *domain.BaseballHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.BaseballHint)
	}
	return nil
}

func (m *MockBaseballPokerGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
