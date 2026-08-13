//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBanLuckGame バンラックゲームモック
type MockBanLuckGame struct {
	mock.Mock
}

func (m *MockBanLuckGame) Reset() { m.Called() }

func (m *MockBanLuckGame) PlaceBet(bet int) error { return m.Called(bet).Error(0) }

func (m *MockBanLuckGame) Hit() error { return m.Called().Error(0) }

func (m *MockBanLuckGame) Stand() error { return m.Called().Error(0) }

func (m *MockBanLuckGame) NextRound() error { return m.Called().Error(0) }

func (m *MockBanLuckGame) CpuPlay() { m.Called() }

func (m *MockBanLuckGame) GetConfig() domain.BanLuckConfig {
	return m.Called().Get(0).(domain.BanLuckConfig)
}

func (m *MockBanLuckGame) SetConfig(cfg domain.BanLuckConfig) { m.Called(cfg) }

func (m *MockBanLuckGame) GetPhase() domain.BanLuckPhase {
	return m.Called().Get(0).(domain.BanLuckPhase)
}

func (m *MockBanLuckGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockBanLuckGame) GetPlayers() []*domain.BanLuckPlayer {
	return m.Called().Get(0).([]*domain.BanLuckPlayer)
}

func (m *MockBanLuckGame) GetHands() []*domain.BlackJackHand {
	return m.Called().Get(0).([]*domain.BlackJackHand)
}

func (m *MockBanLuckGame) GetResults() []domain.BanLuckSeatResult {
	return m.Called().Get(0).([]domain.BanLuckSeatResult)
}

func (m *MockBanLuckGame) GetBankerSeat() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) GetHumanSeat() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockBanLuckGame) MustHit() bool { return m.Called().Bool(0) }

func (m *MockBanLuckGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) WinnerSeat() int { return m.Called().Int(0) }

func (m *MockBanLuckGame) GetHint() *domain.BanLuckHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.BanLuckHint)
	}
	return nil
}

func (m *MockBanLuckGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
