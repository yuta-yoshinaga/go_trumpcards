//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWhistGame ホイストゲームモック
type MockWhistGame struct {
	mock.Mock
}

func (m *MockWhistGame) Reset() {
	m.Called()
}

func (m *MockWhistGame) NextRound() {
	m.Called()
}

func (m *MockWhistGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockWhistGame) CpuPlay() {
	m.Called()
}

func (m *MockWhistGame) ResolveTrick() {
	m.Called()
}

func (m *MockWhistGame) NextTrick() {
	m.Called()
}

func (m *MockWhistGame) ScoreRound() {
	m.Called()
}

func (m *MockWhistGame) GetConfig() domain.WhistConfig {
	args := m.Called()
	return args.Get(0).(domain.WhistConfig)
}

func (m *MockWhistGame) SetConfig(cfg domain.WhistConfig) {
	m.Called(cfg)
}

func (m *MockWhistGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhistGame) GetPhase() domain.WhistPhase {
	args := m.Called()
	return args.Get(0).(domain.WhistPhase)
}

func (m *MockWhistGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWhistGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockWhistGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetTeamScore(team int) int {
	args := m.Called(team)
	return args.Int(0)
}

func (m *MockWhistGame) GetWinnerTeam() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockWhistGame) GetPlayer(i int) *domain.WhistPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.WhistPlayer)
	}
	return nil
}

func (m *MockWhistGame) GetHint() *domain.WhistHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.WhistHint)
	}
	return nil
}

// GetValidPlayIndices はプレイ可能なカードのインデックスを返すモック。
func (m *MockWhistGame) GetValidPlayIndices(playerIdx int) []int {
	out, _ := m.Called(playerIdx).Get(0).([]int)
	return out
}

func (m *MockWhistGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
