//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoofspielGame ゴフスピールゲームモック
type MockGoofspielGame struct {
	mock.Mock
}

func (m *MockGoofspielGame) Reset()   { m.Called() }
func (m *MockGoofspielGame) CpuPlay() { m.Called() }
func (m *MockGoofspielGame) GiveUp()  { m.Called() }

func (m *MockGoofspielGame) PlayerBid(cardIndex int) error { return m.Called(cardIndex).Error(0) }
func (m *MockGoofspielGame) NextRound() error              { return m.Called().Error(0) }

func (m *MockGoofspielGame) GetConfig() domain.GoofspielConfig {
	return m.Called().Get(0).(domain.GoofspielConfig)
}

func (m *MockGoofspielGame) SetConfig(cfg domain.GoofspielConfig) { m.Called(cfg) }

func (m *MockGoofspielGame) GetPhase() domain.GoofspielPhase {
	return m.Called().Get(0).(domain.GoofspielPhase)
}

func (m *MockGoofspielGame) GetGameEndFlag() bool   { return m.Called().Bool(0) }
func (m *MockGoofspielGame) IsHumanTurn() bool      { return m.Called().Bool(0) }
func (m *MockGoofspielGame) HasBid(i int) bool      { return m.Called(i).Bool(0) }
func (m *MockGoofspielGame) PrizeValue() int        { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetPrizeRemaining() int { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetLastWinnerIdx() int  { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetLastGained() int     { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetRoundNumber() int    { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetPlayerCnt() int      { return m.Called().Int(0) }
func (m *MockGoofspielGame) GetWinnerIdx() int      { return m.Called().Int(0) }

func (m *MockGoofspielGame) GetValidBidIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockGoofspielGame) GetCurrentPrize() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockGoofspielGame) GetCarriedPrizes() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockGoofspielGame) GetRevealedBids() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockGoofspielGame) GetPlayer(i int) *domain.GoofspielPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.GoofspielPlayer)
	}
	return nil
}

func (m *MockGoofspielGame) GetHint() *domain.GoofspielHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.GoofspielHint)
	}
	return nil
}

func (m *MockGoofspielGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
