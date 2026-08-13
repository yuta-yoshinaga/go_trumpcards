//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoubleAttackBlackjackGame 追加ベット・ブラックジャックゲームモック
type MockDoubleAttackBlackjackGame struct {
	mock.Mock
}

func (m *MockDoubleAttackBlackjackGame) Reset() { m.Called() }

func (m *MockDoubleAttackBlackjackGame) PlaceBet(ante, bustIt int) error {
	return m.Called(ante, bustIt).Error(0)
}

func (m *MockDoubleAttackBlackjackGame) Attack(amount int) error {
	return m.Called(amount).Error(0)
}

func (m *MockDoubleAttackBlackjackGame) Hit() error { return m.Called().Error(0) }

func (m *MockDoubleAttackBlackjackGame) Stand() error { return m.Called().Error(0) }

func (m *MockDoubleAttackBlackjackGame) Double() error { return m.Called().Error(0) }

func (m *MockDoubleAttackBlackjackGame) Split() error { return m.Called().Error(0) }

func (m *MockDoubleAttackBlackjackGame) NextRound() error { return m.Called().Error(0) }

func (m *MockDoubleAttackBlackjackGame) GetConfig() domain.DoubleAttackBlackjackConfig {
	return m.Called().Get(0).(domain.DoubleAttackBlackjackConfig)
}

func (m *MockDoubleAttackBlackjackGame) SetConfig(cfg domain.DoubleAttackBlackjackConfig) {
	m.Called(cfg)
}

func (m *MockDoubleAttackBlackjackGame) GetPhase() domain.DoubleAttackPhase {
	return m.Called().Get(0).(domain.DoubleAttackPhase)
}

func (m *MockDoubleAttackBlackjackGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockDoubleAttackBlackjackGame) GetHands() []*domain.BlackJackHand {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.BlackJackHand)
}

func (m *MockDoubleAttackBlackjackGame) GetHandCount() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetActiveHandIdx() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetDealerCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockDoubleAttackBlackjackGame) GetDealerScore() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) IsDealerHoleDealt() bool { return m.Called().Bool(0) }

func (m *MockDoubleAttackBlackjackGame) MaxAttackBet() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) CanDouble() bool { return m.Called().Bool(0) }

func (m *MockDoubleAttackBlackjackGame) CanSplit() bool { return m.Called().Bool(0) }

func (m *MockDoubleAttackBlackjackGame) GetAnteBet() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetAttackBet() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetBustItBet() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetResults() []domain.DoubleAttackResult {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]domain.DoubleAttackResult)
}

func (m *MockDoubleAttackBlackjackGame) GetPayout() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetBustItPayout() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetChips() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockDoubleAttackBlackjackGame) GetHint() *domain.DoubleAttackHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.DoubleAttackHint)
}

func (m *MockDoubleAttackBlackjackGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
