//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCincinnatiGame シンシナティゲームモック
type MockCincinnatiGame struct {
	mock.Mock
}

func (m *MockCincinnatiGame) Reset() { m.Called() }

func (m *MockCincinnatiGame) PlayerAction(action, amount int) error {
	return m.Called(action, amount).Error(0)
}

func (m *MockCincinnatiGame) NextHand() error { return m.Called().Error(0) }

func (m *MockCincinnatiGame) CpuPlay() { m.Called() }

func (m *MockCincinnatiGame) GetConfig() domain.CincinnatiConfig {
	return m.Called().Get(0).(domain.CincinnatiConfig)
}

func (m *MockCincinnatiGame) SetConfig(cfg domain.CincinnatiConfig) { m.Called(cfg) }

func (m *MockCincinnatiGame) GetPhase() domain.CincinnatiPhase {
	return m.Called().Get(0).(domain.CincinnatiPhase)
}

func (m *MockCincinnatiGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockCincinnatiGame) GetPlayers() []*domain.CincinnatiPlayer {
	return m.Called().Get(0).([]*domain.CincinnatiPlayer)
}

func (m *MockCincinnatiGame) GetCommunityCards() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}

func (m *MockCincinnatiGame) GetRevealedCount() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetPot() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetCurrentBet() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetToCall() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetRaiseCount() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) CanRaise() bool { return m.Called().Bool(0) }

func (m *MockCincinnatiGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) HumanSeat() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockCincinnatiGame) GetHandNumber() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetResults() []domain.CincinnatiResult {
	return m.Called().Get(0).([]domain.CincinnatiResult)
}

func (m *MockCincinnatiGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockCincinnatiGame) GetHint() *domain.CincinnatiHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.CincinnatiHint)
	}
	return nil
}

func (m *MockCincinnatiGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
