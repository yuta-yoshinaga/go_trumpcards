//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPitchGame ピッチゲームモック
type MockPitchGame struct {
	mock.Mock
}

func (m *MockPitchGame) Reset() {
	m.Called()
}

// GetRoundBreakdown は直近ラウンドの得点内訳を返す。
func (m *MockPitchGame) GetRoundBreakdown() domain.PitchRoundBreakdown {
	args := m.Called()
	bd, _ := args.Get(0).(domain.PitchRoundBreakdown)
	return bd
}

func (m *MockPitchGame) NextRound() {
	m.Called()
}

func (m *MockPitchGame) PlayerBid(bid int) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockPitchGame) CpuBid() {
	m.Called()
}

func (m *MockPitchGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockPitchGame) CpuPlay() {
	m.Called()
}

func (m *MockPitchGame) ResolveTrick() {
	m.Called()
}

func (m *MockPitchGame) NextTrick() {
	m.Called()
}

func (m *MockPitchGame) ScoreRound() {
	m.Called()
}

func (m *MockPitchGame) GetConfig() domain.PitchConfig {
	args := m.Called()
	return args.Get(0).(domain.PitchConfig)
}

func (m *MockPitchGame) SetConfig(cfg domain.PitchConfig) {
	m.Called(cfg)
}

func (m *MockPitchGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPitchGame) GetPhase() domain.PitchPhase {
	args := m.Called()
	return args.Get(0).(domain.PitchPhase)
}

func (m *MockPitchGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPitchGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPitchGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockPitchGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetCurrentBid() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetBidWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPitchGame) GetPlayer(i int) *domain.PitchPlayer {
	args := m.Called(i)
	if v, ok := args.Get(0).(*domain.PitchPlayer); ok {
		return v
	}
	return nil
}

// GetHint モック
func (m *MockPitchGame) GetHint() *domain.PitchHint {
	args := m.Called()
	if v, ok := args.Get(0).(*domain.PitchHint); ok {
		return v
	}
	return nil
}

// GetValidPlayIndices モック
func (m *MockPitchGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

func (m *MockPitchGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
