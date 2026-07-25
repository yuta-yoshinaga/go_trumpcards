//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTwoTenJackGame ツーテンジャックゲームモック
type MockTwoTenJackGame struct {
	mock.Mock
}

func (m *MockTwoTenJackGame) Reset()           { m.Called() }
func (m *MockTwoTenJackGame) NextRound()       { m.Called() }
func (m *MockTwoTenJackGame) CpuDeclareTrump() { m.Called() }
func (m *MockTwoTenJackGame) CpuPlay()         { m.Called() }
func (m *MockTwoTenJackGame) ResolveTrick()    { m.Called() }
func (m *MockTwoTenJackGame) NextTrick()       { m.Called() }
func (m *MockTwoTenJackGame) ScoreRound()      { m.Called() }

func (m *MockTwoTenJackGame) PlayerDeclareTrump(suit int) error {
	args := m.Called(suit)
	return args.Error(0)
}

func (m *MockTwoTenJackGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockTwoTenJackGame) GetConfig() domain.TwoTenJackConfig {
	args := m.Called()
	return args.Get(0).(domain.TwoTenJackConfig)
}

func (m *MockTwoTenJackGame) SetConfig(cfg domain.TwoTenJackConfig) {
	m.Called(cfg)
}

func (m *MockTwoTenJackGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTwoTenJackGame) GetPhase() domain.TwoTenJackPhase {
	args := m.Called()
	return args.Get(0).(domain.TwoTenJackPhase)
}

func (m *MockTwoTenJackGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTwoTenJackGame) IsHumanDeclareTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTwoTenJackGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockTwoTenJackGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetDeclarerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTwoTenJackGame) GetPlayer(i int) *domain.TwoTenJackPlayer {
	args := m.Called(i)
	return args.Get(0).(*domain.TwoTenJackPlayer)
}

func (m *MockTwoTenJackGame) GetHint() *domain.TwoTenJackHint {
	args := m.Called()
	if val, ok := args.Get(0).(*domain.TwoTenJackHint); ok {
		return val
	}
	return nil
}

func (m *MockTwoTenJackGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	return args.Get(0).([]*domain.ActionLogEntry)
}
