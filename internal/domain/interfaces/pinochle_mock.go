//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPinochleGame ピノクルゲームモック
type MockPinochleGame struct {
	mock.Mock
}

func (m *MockPinochleGame) Reset()     { m.Called() }
func (m *MockPinochleGame) NextRound() { m.Called() }

func (m *MockPinochleGame) PlayerBid(amount int) error {
	args := m.Called(amount)
	return args.Error(0)
}

func (m *MockPinochleGame) PlayerPass() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPinochleGame) CpuBid() { m.Called() }

func (m *MockPinochleGame) PlayerCallTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockPinochleGame) CpuCallTrump() { m.Called() }
func (m *MockPinochleGame) ConfirmMelds() { m.Called() }

func (m *MockPinochleGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockPinochleGame) CpuPlay()      { m.Called() }
func (m *MockPinochleGame) ResolveTrick() { m.Called() }
func (m *MockPinochleGame) NextTrick()    { m.Called() }

func (m *MockPinochleGame) GetConfig() domain.PinochleConfig {
	args := m.Called()
	return args.Get(0).(domain.PinochleConfig)
}

func (m *MockPinochleGame) SetConfig(cfg domain.PinochleConfig) { m.Called(cfg) }

func (m *MockPinochleGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPinochleGame) GetPhase() domain.PinochlePhase {
	args := m.Called()
	return args.Get(0).(domain.PinochlePhase)
}

func (m *MockPinochleGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPinochleGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPinochleGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockPinochleGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetHighestBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetHighestBidder() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockPinochleGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPinochleGame) GetPlayer(i int) *domain.PinochlePlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.PinochlePlayer)
}

func (m *MockPinochleGame) GetPlayerMelds() [domain.PinochlePlayerCnt][]*domain.PinochleMeld {
	args := m.Called()
	return args.Get(0).([domain.PinochlePlayerCnt][]*domain.PinochleMeld)
}

func (m *MockPinochleGame) GetHint() *domain.PinochleHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.PinochleHint)
	}
	return nil
}

func (m *MockPinochleGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

func (m *MockPinochleGame) SortHand(playerIdx int) { m.Called(playerIdx) }

func (m *MockPinochleGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
