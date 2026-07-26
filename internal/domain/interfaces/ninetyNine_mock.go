//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNinetyNineGame ナインティナインゲームモック
type MockNinetyNineGame struct {
	mock.Mock
}

func (m *MockNinetyNineGame) Reset() {
	m.Called()
}

func (m *MockNinetyNineGame) NextRound() {
	m.Called()
}

func (m *MockNinetyNineGame) PlayerBid(buryIndices []int) error {
	args := m.Called(buryIndices)
	return args.Error(0)
}

func (m *MockNinetyNineGame) CpuBid() {
	m.Called()
}

func (m *MockNinetyNineGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockNinetyNineGame) CpuPlay() {
	m.Called()
}

func (m *MockNinetyNineGame) ResolveTrick() {
	m.Called()
}

func (m *MockNinetyNineGame) NextTrick() {
	m.Called()
}

func (m *MockNinetyNineGame) ScoreRound() {
	m.Called()
}

func (m *MockNinetyNineGame) GetConfig() domain.NinetyNineConfig {
	args := m.Called()
	return args.Get(0).(domain.NinetyNineConfig)
}

func (m *MockNinetyNineGame) SetConfig(cfg domain.NinetyNineConfig) {
	m.Called(cfg)
}

func (m *MockNinetyNineGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNinetyNineGame) GetPhase() domain.NinetyNinePhase {
	args := m.Called()
	return args.Get(0).(domain.NinetyNinePhase)
}

func (m *MockNinetyNineGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNinetyNineGame) IsHumanBidTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNinetyNineGame) GetDealNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetHandSize() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockNinetyNineGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetBidPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetTargetScore() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNinetyNineGame) GetPlayer(i int) *domain.NinetyNinePlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.NinetyNinePlayer)
}

func (m *MockNinetyNineGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	return args.Get(0).([]int)
}

// GetHint モック
func (m *MockNinetyNineGame) GetHint() *domain.NinetyNineHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.NinetyNineHint); ok {
		return val
	}
	return nil
}

func (m *MockNinetyNineGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}
