//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBriscolaGame ブリスコラゲームモック
type MockBriscolaGame struct {
	mock.Mock
}

func (m *MockBriscolaGame) Reset() {
	m.Called()
}

func (m *MockBriscolaGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBriscolaGame) CpuPlay() {
	m.Called()
}

func (m *MockBriscolaGame) ResolveTrick() {
	m.Called()
}

func (m *MockBriscolaGame) NextTrick() {
	m.Called()
}

func (m *MockBriscolaGame) GetConfig() domain.BriscolaConfig {
	args := m.Called()
	return args.Get(0).(domain.BriscolaConfig)
}

func (m *MockBriscolaGame) SetConfig(cfg domain.BriscolaConfig) {
	m.Called(cfg)
}

func (m *MockBriscolaGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBriscolaGame) GetPhase() domain.BriscolaPhase {
	args := m.Called()
	return args.Get(0).(domain.BriscolaPhase)
}

func (m *MockBriscolaGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBriscolaGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBriscolaGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockBriscolaGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetPlayerPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockBriscolaGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetPlayer(i int) *domain.BriscolaPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BriscolaPlayer)
	}
	return nil
}

func (m *MockBriscolaGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBriscolaGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBriscolaGame) GetHint() *domain.BriscolaHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BriscolaHint)
	}
	return nil
}

func (m *MockBriscolaGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
