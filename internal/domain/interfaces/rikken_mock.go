//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRikkenGame リッケンゲームモック
type MockRikkenGame struct {
	mock.Mock
}

func (m *MockRikkenGame) Reset() {
	m.Called()
}

func (m *MockRikkenGame) Bid(contract int) error {
	args := m.Called(contract)
	return args.Error(0)
}

func (m *MockRikkenGame) Call(trumpSuit int) error {
	args := m.Called(trumpSuit)
	return args.Error(0)
}

func (m *MockRikkenGame) PlayCard(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockRikkenGame) NextRound() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRikkenGame) GiveUp() {
	m.Called()
}

func (m *MockRikkenGame) CpuPlay() {
	m.Called()
}

func (m *MockRikkenGame) GetConfig() domain.RikkenConfig {
	args := m.Called()
	return args.Get(0).(domain.RikkenConfig)
}

func (m *MockRikkenGame) SetConfig(cfg domain.RikkenConfig) {
	m.Called(cfg)
}

func (m *MockRikkenGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRikkenGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRikkenGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockRikkenGame) IsDeclarerSide(playerIdx int) bool {
	args := m.Called(playerIdx)
	return args.Bool(0)
}

func (m *MockRikkenGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetContract() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetPartnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetCalledCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockRikkenGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) HasPassed(i int) bool {
	args := m.Called(i)
	return args.Bool(0)
}

func (m *MockRikkenGame) GetCurrentTurn() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockRikkenGame) GetLastTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockRikkenGame) GetLastTrickWinner() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetTrickCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetDeclarerTricks() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetPlayer(i int) *domain.RikkenPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RikkenPlayer)
}

func (m *MockRikkenGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockRikkenGame) GetHint() *domain.RikkenHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.RikkenHint)
}

func (m *MockRikkenGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
