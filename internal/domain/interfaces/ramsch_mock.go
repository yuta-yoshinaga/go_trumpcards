//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRamschGame Ramsch game mock.
type MockRamschGame struct {
	mock.Mock
}

func (m *MockRamschGame) Reset()     { m.Called() }
func (m *MockRamschGame) NextRound() { m.Called() }

func (m *MockRamschGame) PlayerPlay(cardIndex int) error {
	return m.Called(cardIndex).Error(0)
}
func (m *MockRamschGame) CpuPlay() { m.Called() }

func (m *MockRamschGame) ResolveTrick() { m.Called() }
func (m *MockRamschGame) NextTrick()    { m.Called() }
func (m *MockRamschGame) ScoreRound()   { m.Called() }

func (m *MockRamschGame) GetConfig() domain.RamschConfig {
	return m.Called().Get(0).(domain.RamschConfig)
}
func (m *MockRamschGame) SetConfig(cfg domain.RamschConfig) { m.Called(cfg) }

func (m *MockRamschGame) GetGameEndFlag() bool { return m.Called().Bool(0) }
func (m *MockRamschGame) GetPhase() domain.RamschPhase {
	return m.Called().Get(0).(domain.RamschPhase)
}
func (m *MockRamschGame) IsHumanTurn() bool { return m.Called().Bool(0) }

func (m *MockRamschGame) GetRoundNumber() int      { return m.Called().Int(0) }
func (m *MockRamschGame) GetTrickNumber() int      { return m.Called().Int(0) }
func (m *MockRamschGame) GetCurrentPlayerIdx() int { return m.Called().Int(0) }

func (m *MockRamschGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockRamschGame) GetForehandIdx() int   { return m.Called().Int(0) }
func (m *MockRamschGame) GetMiddlehandIdx() int { return m.Called().Int(0) }
func (m *MockRamschGame) GetRearhandIdx() int   { return m.Called().Int(0) }
func (m *MockRamschGame) GetDealerIdx() int     { return m.Called().Int(0) }

func (m *MockRamschGame) GetLeadPlayerIdx() int { return m.Called().Int(0) }

func (m *MockRamschGame) GetPlayerCnt() int { return m.Called().Int(0) }
func (m *MockRamschGame) GetPlayer(i int) *domain.RamschPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RamschPlayer)
}

func (m *MockRamschGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockRamschGame) GetHint() *domain.RamschHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RamschHint)
}

func (m *MockRamschGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

func (_m *MockRamschGame) GetSkat() []*domain.Card {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).([]*domain.Card)
}

func (_m *MockRamschGame) GetCardPoints(playerIdx int) int {
	ret := _m.Called(playerIdx)
	return ret.Int(0)
}

func (_m *MockRamschGame) GetLoserIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRamschGame) IsDurchmarsch() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRamschGame) GetDurchmarschIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}
