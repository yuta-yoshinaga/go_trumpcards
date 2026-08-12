//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrazyFourPokerGame クレイジー 4 ポーカーゲームモック
type MockCrazyFourPokerGame struct {
	mock.Mock
}

func (m *MockCrazyFourPokerGame) Reset() { m.Called() }

func (m *MockCrazyFourPokerGame) PlaceBet(ante, queensUp int) error {
	return m.Called(ante, queensUp).Error(0)
}

func (m *MockCrazyFourPokerGame) Play(multiplier int) error {
	return m.Called(multiplier).Error(0)
}

func (m *MockCrazyFourPokerGame) Fold() error { return m.Called().Error(0) }

func (m *MockCrazyFourPokerGame) NextRound() error { return m.Called().Error(0) }

func (m *MockCrazyFourPokerGame) GetConfig() domain.CrazyFourPokerConfig {
	return m.Called().Get(0).(domain.CrazyFourPokerConfig)
}

func (m *MockCrazyFourPokerGame) SetConfig(cfg domain.CrazyFourPokerConfig) { m.Called(cfg) }

func (m *MockCrazyFourPokerGame) GetPhase() domain.CrazyFourPokerPhase {
	return m.Called().Get(0).(domain.CrazyFourPokerPhase)
}

func (m *MockCrazyFourPokerGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockCrazyFourPokerGame) GetPlayerHand() []*domain.Card { return mockCards(m.Called()) }

func (m *MockCrazyFourPokerGame) GetDealerHand() []*domain.Card { return mockCards(m.Called()) }

func (m *MockCrazyFourPokerGame) GetPlayerBest() []*domain.Card { return mockCards(m.Called()) }

func (m *MockCrazyFourPokerGame) GetDealerBest() []*domain.Card { return mockCards(m.Called()) }

// mockCards は nil を安全に返すための共通処理。
func mockCards(args mock.Arguments) []*domain.Card {
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockCrazyFourPokerGame) GetPlayerHandRank() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetDealerHandRank() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) PlayerHasAcesOrBetter() bool { return m.Called().Bool(0) }

func (m *MockCrazyFourPokerGame) MaxPlayMultiplier() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) PlayerQualifies() bool { return m.Called().Bool(0) }

func (m *MockCrazyFourPokerGame) DealerQualifies() bool { return m.Called().Bool(0) }

func (m *MockCrazyFourPokerGame) GetAnteBet() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetSuperBet() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetQueensUpBet() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetPlayBet() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetPlayMultiplier() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetResult() domain.CrazyFourPokerResult {
	return m.Called().Get(0).(domain.CrazyFourPokerResult)
}

func (m *MockCrazyFourPokerGame) GetPayout() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetChips() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetMinTotalWager() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetRoundNumber() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetRemainingCards() int { return m.Called().Int(0) }

func (m *MockCrazyFourPokerGame) GetHint() *domain.CrazyFourPokerHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.CrazyFourPokerHint)
}

func (m *MockCrazyFourPokerGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
