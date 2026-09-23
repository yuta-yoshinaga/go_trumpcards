//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBinokelGame ビノクルゲームモック
type MockBinokelGame struct {
	mock.Mock
}

func (m *MockBinokelGame) Reset()     { m.Called() }
func (m *MockBinokelGame) NextRound() { m.Called() }

func (m *MockBinokelGame) PlayerBid(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockBinokelGame) PlayerPass() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBinokelGame) CpuBid() { m.Called() }

func (m *MockBinokelGame) PlayerDiscardToDabb(cardIndices []int) error {
	args := m.Called(cardIndices)
	return args.Error(0)
}

func (m *MockBinokelGame) CpuDiscardToDabb() { m.Called() }

func (m *MockBinokelGame) IsHumanDabbTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBinokelGame) PlayerCallTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockBinokelGame) CpuCallTrump() { m.Called() }
func (m *MockBinokelGame) ConfirmMelds() { m.Called() }

func (m *MockBinokelGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBinokelGame) CpuPlay()      { m.Called() }
func (m *MockBinokelGame) ResolveTrick() { m.Called() }
func (m *MockBinokelGame) NextTrick()    { m.Called() }

func (m *MockBinokelGame) GetConfig() domain.BinokelConfig {
	args := m.Called()
	return args.Get(0).(domain.BinokelConfig)
}

func (m *MockBinokelGame) SetConfig(cfg domain.BinokelConfig) { m.Called(cfg) }

func (m *MockBinokelGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBinokelGame) GetPhase() domain.BinokelPhase {
	args := m.Called()
	return args.Get(0).(domain.BinokelPhase)
}

func (m *MockBinokelGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBinokelGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBinokelGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockBinokelGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetHighestBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetHighestBidder() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetScores() [domain.BinokelPlayerCnt]int {
	args := m.Called()
	return args.Get(0).([domain.BinokelPlayerCnt]int)
}

func (m *MockBinokelGame) GetScore(playerIdx int) int {
	args := m.Called(playerIdx)
	return args.Int(0)
}

func (m *MockBinokelGame) GetWinnerPlayer() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetDabb() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockBinokelGame) GetDabbDiscarded() []*domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

func (m *MockBinokelGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBinokelGame) GetPlayer(i int) *domain.BinokelPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.BinokelPlayer)
}

func (m *MockBinokelGame) GetPlayerMelds() [domain.BinokelPlayerCnt][]*domain.BinokelMeld {
	args := m.Called()
	return args.Get(0).([domain.BinokelPlayerCnt][]*domain.BinokelMeld)
}

func (m *MockBinokelGame) GetHint() *domain.BinokelHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BinokelHint)
	}
	return nil
}

func (m *MockBinokelGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

func (m *MockBinokelGame) SortHand(playerIdx int) { m.Called(playerIdx) }

func (m *MockBinokelGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
