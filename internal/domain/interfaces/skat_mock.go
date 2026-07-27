//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSkatGame Skat game mock.
type MockSkatGame struct {
	mock.Mock
}

func (m *MockSkatGame) Reset()     { m.Called() }
func (m *MockSkatGame) NextRound() { m.Called() }

func (m *MockSkatGame) PlayerBid(accept bool) error {
	return m.Called(accept).Error(0)
}
func (m *MockSkatGame) CpuBid() { m.Called() }

func (m *MockSkatGame) PlayerPickSkat(pickup bool) error {
	return m.Called(pickup).Error(0)
}
func (m *MockSkatGame) CpuPickSkat() { m.Called() }

func (m *MockSkatGame) PlayerDiscard(idxA, idxB int) error {
	return m.Called(idxA, idxB).Error(0)
}
func (m *MockSkatGame) CpuDiscard() { m.Called() }

func (m *MockSkatGame) PlayerDeclareGame(gameType domain.SkatGameType, trumpSuit int) error {
	return m.Called(gameType, trumpSuit).Error(0)
}
func (m *MockSkatGame) CpuDeclareGame() { m.Called() }

func (m *MockSkatGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockSkatGame) CpuPlay() { m.Called() }

func (m *MockSkatGame) ResolveTrick() { m.Called() }
func (m *MockSkatGame) NextTrick()    { m.Called() }
func (m *MockSkatGame) ScoreRound()   { m.Called() }

func (m *MockSkatGame) GetConfig() domain.SkatConfig {
	return m.Called().Get(0).(domain.SkatConfig)
}
func (m *MockSkatGame) SetConfig(cfg domain.SkatConfig) { m.Called(cfg) }

func (m *MockSkatGame) GetGameEndFlag() bool { return m.Called().Bool(0) }
func (m *MockSkatGame) GetPhase() domain.SkatPhase {
	return m.Called().Get(0).(domain.SkatPhase)
}
func (m *MockSkatGame) IsHumanTurn() bool         { return m.Called().Bool(0) }
func (m *MockSkatGame) IsHumanBidTurn() bool      { return m.Called().Bool(0) }
func (m *MockSkatGame) IsHumanDeclarerTurn() bool { return m.Called().Bool(0) }

func (m *MockSkatGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockSkatGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockSkatGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }

func (m *MockSkatGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockSkatGame) GetForehandIdx() int       { return m.Called().Int(0) }
func (m *MockSkatGame) GetMiddlehandIdx() int     { return m.Called().Int(0) }
func (m *MockSkatGame) GetRearhandIdx() int       { return m.Called().Int(0) }
func (m *MockSkatGame) GetDealerIdx() int         { return m.Called().Int(0) }
func (m *MockSkatGame) GetDeclarerIdx() int       { return m.Called().Int(0) }
func (m *MockSkatGame) GetCurrentBid() int        { return m.Called().Int(0) }
func (m *MockSkatGame) GetActiveBidActorIdx() int { return m.Called().Int(0) }

func (m *MockSkatGame) GetGameType() domain.SkatGameType {
	return m.Called().Get(0).(domain.SkatGameType)
}
func (m *MockSkatGame) GetTrumpSuit() int { return m.Called().Int(0) }

func (m *MockSkatGame) GetSkat() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}
func (m *MockSkatGame) GetOriginalSkat() []*domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.Card)
}

func (m *MockSkatGame) GetDeclarerCardPoints() int  { return m.Called().Int(0) }
func (m *MockSkatGame) GetDefendersCardPoints() int { return m.Called().Int(0) }
func (m *MockSkatGame) GetWinnerSide() int          { return m.Called().Int(0) }
func (m *MockSkatGame) GetGameValue() int           { return m.Called().Int(0) }
func (m *MockSkatGame) GetLeadPlayerIdx() int       { return m.Called().Int(0) }
func (m *MockSkatGame) PickedSkat() bool            { return m.Called().Bool(0) }

func (m *MockSkatGame) GetPlayerCnt() int { return m.Called().Int(0) }
func (m *MockSkatGame) GetPlayer(i int) *domain.SkatPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.SkatPlayer)
}

func (m *MockSkatGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockSkatGame) GetHint() *domain.SkatHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.SkatHint)
}

func (m *MockSkatGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
