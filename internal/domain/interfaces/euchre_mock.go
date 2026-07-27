//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEuchreGame ユーカーゲームモック
type MockEuchreGame struct {
	mock.Mock
}

func (m *MockEuchreGame) Reset()     { m.Called() }
func (m *MockEuchreGame) NextRound() { m.Called() }

func (m *MockEuchreGame) PlayerPickUp(orderUp bool, goAlone bool) error {
	args := m.Called(orderUp, goAlone)
	return args.Error(0)
}

func (m *MockEuchreGame) CpuPickUp() { m.Called() }

func (m *MockEuchreGame) PlayerCallTrump(suit int, goAlone bool) error {
	args := m.Called(suit, goAlone)
	return args.Error(0)
}

func (m *MockEuchreGame) PlayerPassCall() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEuchreGame) CpuCallTrump() { m.Called() }

func (m *MockEuchreGame) PlayerDiscard(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockEuchreGame) CpuDiscard() { m.Called() }

func (m *MockEuchreGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockEuchreGame) CpuPlay()      { m.Called() }
func (m *MockEuchreGame) ResolveTrick() { m.Called() }
func (m *MockEuchreGame) NextTrick()    { m.Called() }
func (m *MockEuchreGame) ScoreRound()   { m.Called() }

func (m *MockEuchreGame) GetConfig() domain.EuchreConfig {
	args := m.Called()
	return args.Get(0).(domain.EuchreConfig)
}

func (m *MockEuchreGame) SetConfig(cfg domain.EuchreConfig) { m.Called(cfg) }

func (m *MockEuchreGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEuchreGame) GetPhase() domain.EuchrePhase {
	args := m.Called()
	return args.Get(0).(domain.EuchrePhase)
}

func (m *MockEuchreGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEuchreGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEuchreGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockEuchreGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetFaceUpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockEuchreGame) GetMakerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetGoingAlone() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEuchreGame) GetGoingAlonePlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockEuchreGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEuchreGame) GetPlayer(i int) *domain.EuchrePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.EuchrePlayer)
	}
	return nil
}

func (m *MockEuchreGame) GetKitty() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockEuchreGame) GetHint() *domain.EuchreHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.EuchreHint)
	}
	return nil
}

func (m *MockEuchreGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

func (m *MockEuchreGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}
