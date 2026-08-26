//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpeculationGame はスペキュレーションゲームモック
type MockSpeculationGame struct {
	mock.Mock
}

func (m *MockSpeculationGame) Reset() { m.Called() }

func (m *MockSpeculationGame) Flip() error { return m.Called().Error(0) }

func (m *MockSpeculationGame) Accept() error { return m.Called().Error(0) }

func (m *MockSpeculationGame) Decline() error { return m.Called().Error(0) }

func (m *MockSpeculationGame) Bid(amount int) error { return m.Called(amount).Error(0) }

func (m *MockSpeculationGame) NextRound() error { return m.Called().Error(0) }

func (m *MockSpeculationGame) GetPhase() domain.SpeculationPhase {
	return m.Called().Get(0).(domain.SpeculationPhase)
}

func (m *MockSpeculationGame) GetPlayers() []*domain.SpeculationPlayer {
	args := m.Called()
	if v, ok := args.Get(0).([]*domain.SpeculationPlayer); ok {
		return v
	}
	return nil
}

func (m *MockSpeculationGame) GetConfig() domain.SpeculationConfig {
	return m.Called().Get(0).(domain.SpeculationConfig)
}

func (m *MockSpeculationGame) GetTrumpSuit() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v, ok := args.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (m *MockSpeculationGame) GetPot() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetTurnSeat() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetBestSeat() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetOfferFrom() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetOfferTo() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetOfferAmount() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetRoundNo() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetWinnerSeat() int { return m.Called().Int(0) }

func (m *MockSpeculationGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockSpeculationGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
