//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockIronCrossGame アイアンクロスゲームモック
type MockIronCrossGame struct {
	mock.Mock
}

func (m *MockIronCrossGame) Reset() { m.Called() }

func (m *MockIronCrossGame) PlayerAction(action, amount int) error {
	return m.Called(action, amount).Error(0)
}

func (m *MockIronCrossGame) ChooseLine(l domain.IronCrossLine) error {
	return m.Called(l).Error(0)
}

func (m *MockIronCrossGame) NextHand() error { return m.Called().Error(0) }

func (m *MockIronCrossGame) CpuPlay() { m.Called() }

func (m *MockIronCrossGame) GetConfig() domain.IronCrossConfig {
	return m.Called().Get(0).(domain.IronCrossConfig)
}

func (m *MockIronCrossGame) SetConfig(cfg domain.IronCrossConfig) { m.Called(cfg) }

func (m *MockIronCrossGame) GetPhase() domain.IronCrossPhase {
	return m.Called().Get(0).(domain.IronCrossPhase)
}

func (m *MockIronCrossGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockIronCrossGame) GetPlayers() []*domain.IronCrossPlayer {
	return m.Called().Get(0).([]*domain.IronCrossPlayer)
}

func (m *MockIronCrossGame) GetCross() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}

func (m *MockIronCrossGame) GetRevealedCount() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetPot() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetCurrentBet() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetToCall() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetRaiseCount() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) CanRaise() bool { return m.Called().Bool(0) }

func (m *MockIronCrossGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) HumanSeat() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockIronCrossGame) IsChoosing() bool { return m.Called().Bool(0) }

func (m *MockIronCrossGame) GetHandNumber() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetResults() []domain.IronCrossResult {
	return m.Called().Get(0).([]domain.IronCrossResult)
}

func (m *MockIronCrossGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockIronCrossGame) GetHint() *domain.IronCrossHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.IronCrossHint)
	}
	return nil
}

func (m *MockIronCrossGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
