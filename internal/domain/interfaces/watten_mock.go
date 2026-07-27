//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWattenGame ヴァッテンゲームモック
type MockWattenGame struct {
	mock.Mock
}

func (m *MockWattenGame) Reset()     { m.Called() }
func (m *MockWattenGame) NextRound() { m.Called() }

func (m *MockWattenGame) PlayerDeclare(rank, suit int) error {
	args := m.Called(rank, suit)
	return args.Error(0)
}

func (m *MockWattenGame) CpuDeclare() { m.Called() }

func (m *MockWattenGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockWattenGame) CpuPlay() { m.Called() }

func (m *MockWattenGame) PlayerRaise() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWattenGame) PlayerRespond(hold bool) error {
	args := m.Called(hold)
	return args.Error(0)
}

func (m *MockWattenGame) CpuRespond()   { m.Called() }
func (m *MockWattenGame) ResolveTrick() { m.Called() }
func (m *MockWattenGame) NextTrick()    { m.Called() }
func (m *MockWattenGame) ScoreRound()   { m.Called() }

func (m *MockWattenGame) GetConfig() domain.WattenConfig {
	args := m.Called()
	return args.Get(0).(domain.WattenConfig)
}

func (m *MockWattenGame) SetConfig(cfg domain.WattenConfig) { m.Called(cfg) }

func (m *MockWattenGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWattenGame) GetPhase() domain.WattenPhase {
	args := m.Called()
	return args.Get(0).(domain.WattenPhase)
}

func (m *MockWattenGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWattenGame) IsHumanDeclareTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWattenGame) IsHumanRespondTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWattenGame) CanHumanRaise() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWattenGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockWattenGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetSchlagRank() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetCriticalSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetStake() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetPendingStake() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetRaiseCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetRaiserTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetResponderIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockWattenGame) GetTeamTricks(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockWattenGame) GetDealWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetResult() domain.WattenResult {
	args := m.Called()
	return args.Get(0).(domain.WattenResult)
}

func (m *MockWattenGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWattenGame) GetPlayer(i int) *domain.WattenPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.WattenPlayer)
	}
	return nil
}

func (m *MockWattenGame) GetHint() *domain.WattenHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.WattenHint)
	}
	return nil
}

func (m *MockWattenGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockWattenGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}
