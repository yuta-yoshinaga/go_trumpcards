//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockColourWhistGame カラーホイストゲームモック
type MockColourWhistGame struct {
	mock.Mock
}

func (m *MockColourWhistGame) Reset() {
	m.Called()
}

func (m *MockColourWhistGame) Bid(contract int) error {
	args := m.Called(contract)
	return args.Error(0)
}

func (m *MockColourWhistGame) Call(trumpSuit int) error {
	args := m.Called(trumpSuit)
	return args.Error(0)
}

func (m *MockColourWhistGame) PlayCard(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockColourWhistGame) NextRound() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockColourWhistGame) GiveUp() {
	m.Called()
}

func (m *MockColourWhistGame) CpuPlay() {
	m.Called()
}

func (m *MockColourWhistGame) GetConfig() domain.ColourWhistConfig {
	args := m.Called()
	return args.Get(0).(domain.ColourWhistConfig)
}

func (m *MockColourWhistGame) SetConfig(cfg domain.ColourWhistConfig) {
	m.Called(cfg)
}

func (m *MockColourWhistGame) GetPhase() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockColourWhistGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockColourWhistGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockColourWhistGame) IsDeclarerSide(playerIdx int) bool {
	args := m.Called(playerIdx)
	return args.Bool(0)
}

func (m *MockColourWhistGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetContract() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetPartnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetCalledCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockColourWhistGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) IsTroelForced() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockColourWhistGame) HasPassed(i int) bool {
	args := m.Called(i)
	return args.Bool(0)
}

func (m *MockColourWhistGame) GetCurrentTurn() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockColourWhistGame) GetLastTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockColourWhistGame) GetLastTrickWinner() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetTrickCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetDeclarerTricks() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetPlayer(i int) *domain.ColourWhistPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ColourWhistPlayer)
}

func (m *MockColourWhistGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockColourWhistGame) GetHint() *domain.ColourWhistHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ColourWhistHint)
}

func (m *MockColourWhistGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}

func (m *MockColourWhistGame) IsDeclarerSideVisible(playerIdx int) bool {
	args := m.Called(playerIdx)
	return args.Bool(0)
}
