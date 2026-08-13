//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFreeBetBlackjackGame フリーベット・ブラックジャックゲームモック
type MockFreeBetBlackjackGame struct {
	mock.Mock
}

func (m *MockFreeBetBlackjackGame) Reset() { m.Called() }

func (m *MockFreeBetBlackjackGame) PlaceBet(ante int) error { return m.Called(ante).Error(0) }

func (m *MockFreeBetBlackjackGame) Hit() error { return m.Called().Error(0) }

func (m *MockFreeBetBlackjackGame) Stand() error { return m.Called().Error(0) }

func (m *MockFreeBetBlackjackGame) FreeDouble() error { return m.Called().Error(0) }

func (m *MockFreeBetBlackjackGame) FreeSplit() error { return m.Called().Error(0) }

func (m *MockFreeBetBlackjackGame) NextRound() error { return m.Called().Error(0) }

func (m *MockFreeBetBlackjackGame) GetConfig() domain.FreeBetBlackjackConfig {
	return m.Called().Get(0).(domain.FreeBetBlackjackConfig)
}

func (m *MockFreeBetBlackjackGame) SetConfig(cfg domain.FreeBetBlackjackConfig) { m.Called(cfg) }

func (m *MockFreeBetBlackjackGame) GetPhase() domain.FreeBetPhase {
	return m.Called().Get(0).(domain.FreeBetPhase)
}

func (m *MockFreeBetBlackjackGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockFreeBetBlackjackGame) GetHands() []*domain.BlackJackHand {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.BlackJackHand)
}

func (m *MockFreeBetBlackjackGame) GetHandCount() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetFreeBets() []int {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockFreeBetBlackjackGame) GetFreeBet(idx int) int { return m.Called(idx).Int(0) }

func (m *MockFreeBetBlackjackGame) GetActiveHandIdx() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetDealerCards() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockFreeBetBlackjackGame) GetDealerScore() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) IsDealerPushed22() bool { return m.Called().Bool(0) }

func (m *MockFreeBetBlackjackGame) CanFreeDouble() bool { return m.Called().Bool(0) }

func (m *MockFreeBetBlackjackGame) CanFreeSplit() bool { return m.Called().Bool(0) }

func (m *MockFreeBetBlackjackGame) GetAnteBet() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetResults() []domain.FreeBetResult {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]domain.FreeBetResult)
}

func (m *MockFreeBetBlackjackGame) GetPayout() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetChips() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockFreeBetBlackjackGame) GetHint() *domain.FreeBetHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.FreeBetHint)
}

func (m *MockFreeBetBlackjackGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
