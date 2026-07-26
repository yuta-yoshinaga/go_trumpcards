//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOhHellGame オー・ヘルゲームモック
type MockOhHellGame struct {
	mock.Mock
}

func (m *MockOhHellGame) Reset() {
	m.Called()
}

func (m *MockOhHellGame) NextRound() {
	m.Called()
}

func (m *MockOhHellGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockOhHellGame) CpuBid() {
	m.Called()
}

func (m *MockOhHellGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockOhHellGame) CpuPlay() {
	m.Called()
}

func (m *MockOhHellGame) ResolveTrick() {
	m.Called()
}

func (m *MockOhHellGame) NextTrick() {
	m.Called()
}

func (m *MockOhHellGame) ScoreRound() {
	m.Called()
}

func (m *MockOhHellGame) GetConfig() domain.OhHellConfig {
	args := m.Called()
	return args.Get(0).(domain.OhHellConfig)
}

func (m *MockOhHellGame) SetConfig(cfg domain.OhHellConfig) {
	m.Called(cfg)
}

func (m *MockOhHellGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOhHellGame) GetPhase() domain.OhHellPhase {
	args := m.Called()
	return args.Get(0).(domain.OhHellPhase)
}

func (m *MockOhHellGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOhHellGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOhHellGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetTotalRounds() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetHandSize() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockOhHellGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

func (m *MockOhHellGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetRestrictedBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockOhHellGame) GetPlayer(i int) *domain.OhHellPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.OhHellPlayer)
}

func (m *MockOhHellGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

// GetHint モック
func (m *MockOhHellGame) GetHint() *domain.OhHellHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.OhHellHint); ok {
		return val
	}
	return nil
}

func (m *MockOhHellGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}
