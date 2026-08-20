//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMendikotGame メンディコットゲームモック
type MockMendikotGame struct {
	mock.Mock
}

func (m *MockMendikotGame) Reset()    { m.Called() }
func (m *MockMendikotGame) CpuPlay()  { m.Called() }
func (m *MockMendikotGame) NextHand() { m.Called() }
func (m *MockMendikotGame) GiveUp()   { m.Called() }

func (m *MockMendikotGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}

func (m *MockMendikotGame) GetConfig() domain.MendikotConfig {
	return m.Called().Get(0).(domain.MendikotConfig)
}

func (m *MockMendikotGame) SetConfig(cfg domain.MendikotConfig) { m.Called(cfg) }

func (m *MockMendikotGame) GetGameEndFlag() bool { return m.Called().Bool(0) }

func (m *MockMendikotGame) GetPhase() domain.MendikotPhase {
	return m.Called().Get(0).(domain.MendikotPhase)
}

func (m *MockMendikotGame) IsHumanTurn() bool        { return m.Called().Bool(0) }
func (m *MockMendikotGame) GetHandNumber() int       { return m.Called().Int(0) }
func (m *MockMendikotGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockMendikotGame) GetTrumpSuit() int        { return m.Called().Int(0) }
func (m *MockMendikotGame) GetTrumpChooserIdx() int  { return m.Called().Int(0) }
func (m *MockMendikotGame) GetLastHandWinner() int   { return m.Called().Int(0) }
func (m *MockMendikotGame) GetLastHandKind() string  { return m.Called().String(0) }
func (m *MockMendikotGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }
func (m *MockMendikotGame) GetLeadPlayerIdx() int    { return m.Called().Int(0) }
func (m *MockMendikotGame) GetDealerIdx() int        { return m.Called().Int(0) }
func (m *MockMendikotGame) GetPlayerCnt() int        { return m.Called().Int(0) }
func (m *MockMendikotGame) GetWinnerTeam() int       { return m.Called().Int(0) }

func (m *MockMendikotGame) GetScore(team int) int   { return m.Called(team).Int(0) }
func (m *MockMendikotGame) TeamTens(team int) int   { return m.Called(team).Int(0) }
func (m *MockMendikotGame) TeamTricks(team int) int { return m.Called(team).Int(0) }

func (m *MockMendikotGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

// WillSetTrump モック
func (m *MockMendikotGame) WillSetTrump(playerIdx int) bool {
	return m.Called(playerIdx).Bool(0)
}

func (m *MockMendikotGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockMendikotGame) GetPlayer(i int) *domain.MendikotPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.MendikotPlayer)
}

func (m *MockMendikotGame) GetHint() *domain.MendikotHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.MendikotHint)
}

func (m *MockMendikotGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
