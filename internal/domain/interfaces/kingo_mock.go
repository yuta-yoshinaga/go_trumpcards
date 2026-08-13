//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKingoGame キンゴゲームモック
type MockKingoGame struct {
	mock.Mock
}

func (m *MockKingoGame) Reset() { m.Called() }

func (m *MockKingoGame) PlaceBet(amount int) error { return m.Called(amount).Error(0) }

func (m *MockKingoGame) Deal() error { return m.Called().Error(0) }

func (m *MockKingoGame) NextRound() error { return m.Called().Error(0) }

func (m *MockKingoGame) GetConfig() domain.KingoConfig {
	return m.Called().Get(0).(domain.KingoConfig)
}

func (m *MockKingoGame) SetConfig(cfg domain.KingoConfig) { m.Called(cfg) }

func (m *MockKingoGame) GetPhase() domain.KingoPhase {
	return m.Called().Get(0).(domain.KingoPhase)
}

func (m *MockKingoGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockKingoGame) GetPlayers() []*domain.KingoPlayer {
	return m.Called().Get(0).([]*domain.KingoPlayer)
}

func (m *MockKingoGame) GetBankerSeat() int { return m.Called().Int(0) }

func (m *MockKingoGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockKingoGame) GetResults() []domain.KingoResult {
	return m.Called().Get(0).([]domain.KingoResult)
}

func (m *MockKingoGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockKingoGame) HumanSeat() int { return m.Called().Int(0) }

func (m *MockKingoGame) IsHumanBanker() bool { return m.Called().Bool(0) }

func (m *MockKingoGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockKingoGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockKingoGame) GetHint() *domain.KingoHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.KingoHint)
	}
	return nil
}

func (m *MockKingoGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
