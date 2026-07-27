//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAllFoursGame All Fours ゲームモック
type MockAllFoursGame struct {
	mock.Mock
}

func (m *MockAllFoursGame) Reset() {
	m.Called()
}

func (m *MockAllFoursGame) NextRound() {
	m.Called()
}

func (m *MockAllFoursGame) PlayerBeg(beg bool) error {
	args := m.Called(beg)
	return args.Error(0)
}

func (m *MockAllFoursGame) CpuBeg() {
	m.Called()
}

func (m *MockAllFoursGame) PlayerRespondBeg(run bool) error {
	args := m.Called(run)
	return args.Error(0)
}

func (m *MockAllFoursGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockAllFoursGame) CpuPlay() {
	m.Called()
}

func (m *MockAllFoursGame) ResolveTrick() {
	m.Called()
}

func (m *MockAllFoursGame) NextTrick() {
	m.Called()
}

func (m *MockAllFoursGame) ScoreRound() {
	m.Called()
}

func (m *MockAllFoursGame) GetConfig() domain.AllFoursConfig {
	args := m.Called()
	return args.Get(0).(domain.AllFoursConfig)
}

func (m *MockAllFoursGame) SetConfig(cfg domain.AllFoursConfig) {
	m.Called(cfg)
}

func (m *MockAllFoursGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAllFoursGame) GetPhase() domain.AllFoursPhase {
	args := m.Called()
	return args.Get(0).(domain.AllFoursPhase)
}

func (m *MockAllFoursGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAllFoursGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetNonDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

func (m *MockAllFoursGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetTurnUp() *domain.Card {
	args := m.Called()
	if v, ok := args.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

func (m *MockAllFoursGame) GetRunCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockAllFoursGame) GetPlayer(i int) *domain.AllFoursPlayer {
	args := m.Called(i)
	if v, ok := args.Get(0).(*domain.AllFoursPlayer); ok {
		return v
	}
	return nil
}

// GetHint モック
func (m *MockAllFoursGame) GetHint() *domain.AllFoursHint {
	args := m.Called()
	if v, ok := args.Get(0).(*domain.AllFoursHint); ok {
		return v
	}
	return nil
}

// GetValidPlayIndices モック
func (m *MockAllFoursGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

func (m *MockAllFoursGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
