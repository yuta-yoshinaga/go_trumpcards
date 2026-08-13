//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMonteBankGame モンテバンクゲームモック
type MockMonteBankGame struct {
	mock.Mock
}

func (m *MockMonteBankGame) Reset() { m.Called() }

func (m *MockMonteBankGame) PlaceBet(idx, bet int) error { return m.Called(idx, bet).Error(0) }

func (m *MockMonteBankGame) NextRound() error { return m.Called().Error(0) }

func (m *MockMonteBankGame) GetConfig() domain.MonteBankConfig {
	return m.Called().Get(0).(domain.MonteBankConfig)
}

func (m *MockMonteBankGame) SetConfig(cfg domain.MonteBankConfig) { m.Called(cfg) }

func (m *MockMonteBankGame) GetPhase() domain.MonteBankPhase {
	return m.Called().Get(0).(domain.MonteBankPhase)
}

func (m *MockMonteBankGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockMonteBankGame) GetLayout() []*domain.Card {
	return m.Called().Get(0).([]*domain.Card)
}

func (m *MockMonteBankGame) GetGate() *domain.Card {
	if c := m.Called().Get(0); c != nil {
		return c.(*domain.Card)
	}
	return nil
}

func (m *MockMonteBankGame) GetPick() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) GetBet() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) GetResult() domain.MonteBankResult {
	return m.Called().Get(0).(domain.MonteBankResult)
}

func (m *MockMonteBankGame) GetPayout() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) SuitCountInLayout(design int) int { return m.Called(design).Int(0) }

func (m *MockMonteBankGame) RemainingOfSuit(design int) int { return m.Called(design).Int(0) }

func (m *MockMonteBankGame) GetChips() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockMonteBankGame) GetHint() *domain.MonteBankHint {
	if h := m.Called().Get(0); h != nil {
		return h.(*domain.MonteBankHint)
	}
	return nil
}

func (m *MockMonteBankGame) GetActionLog() []*domain.ActionLogEntry {
	return m.Called().Get(0).([]*domain.ActionLogEntry)
}
