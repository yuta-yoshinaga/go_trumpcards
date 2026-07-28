//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeziqueGame ベジークゲームモック
type MockBeziqueGame struct {
	mock.Mock
}

func (m *MockBeziqueGame) Reset() {
	m.Called()
}

func (m *MockBeziqueGame) NextRound() {
	m.Called()
}

func (m *MockBeziqueGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBeziqueGame) CpuPlay() {
	m.Called()
}

func (m *MockBeziqueGame) PlayerDeclareMeld(meldIndex int) error {
	args := m.Called(meldIndex)
	return args.Error(0)
}

func (m *MockBeziqueGame) PlayerSkipMeld() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBeziqueGame) CpuMeld() {
	m.Called()
}

func (m *MockBeziqueGame) GetConfig() domain.BeziqueConfig {
	args := m.Called()
	return args.Get(0).(domain.BeziqueConfig)
}

func (m *MockBeziqueGame) SetConfig(cfg domain.BeziqueConfig) {
	m.Called(cfg)
}

func (m *MockBeziqueGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeziqueGame) GetPhase() domain.BeziquePhase {
	args := m.Called()
	return args.Get(0).(domain.BeziquePhase)
}

func (m *MockBeziqueGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeziqueGame) IsEndgame() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBeziqueGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBeziqueGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockBeziqueGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetDealPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

// GetDealMeldPoints モック
func (m *MockBeziqueGame) GetDealMeldPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockBeziqueGame) GetMatchScore(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockBeziqueGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetPlayer(i int) *domain.BeziquePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BeziquePlayer)
	}
	return nil
}

func (m *MockBeziqueGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBeziqueGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBeziqueGame) GetAvailableMelds(playerIdx int) []domain.BeziqueMeld {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]domain.BeziqueMeld)
	}
	return nil
}

func (m *MockBeziqueGame) GetHint() *domain.BeziqueHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BeziqueHint)
	}
	return nil
}

func (m *MockBeziqueGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
