//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEcarteGame エカルテゲームモック
type MockEcarteGame struct {
	mock.Mock
}

func (m *MockEcarteGame) Reset() {
	m.Called()
}

func (m *MockEcarteGame) NextRound() {
	m.Called()
}

func (m *MockEcarteGame) PlayerPropose() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEcarteGame) PlayerStand() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEcarteGame) PlayerRespond(accept bool) error {
	args := m.Called(accept)
	return args.Error(0)
}

func (m *MockEcarteGame) PlayerDiscard(indices []int) error {
	args := m.Called(indices)
	return args.Error(0)
}

func (m *MockEcarteGame) CpuExchange() {
	m.Called()
}

func (m *MockEcarteGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockEcarteGame) CpuPlay() {
	m.Called()
}

func (m *MockEcarteGame) GetConfig() domain.EcarteConfig {
	args := m.Called()
	return args.Get(0).(domain.EcarteConfig)
}

func (m *MockEcarteGame) SetConfig(cfg domain.EcarteConfig) {
	m.Called(cfg)
}

func (m *MockEcarteGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEcarteGame) GetPhase() domain.EcartePhase {
	args := m.Called()
	return args.Get(0).(domain.EcartePhase)
}

func (m *MockEcarteGame) GetNegStep() domain.EcarteNegStep {
	args := m.Called()
	return args.Get(0).(domain.EcarteNegStep)
}

func (m *MockEcarteGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEcarteGame) GetRoundNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockEcarteGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockEcarteGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetElderIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) IsRefusalByDealer() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockEcarteGame) GetDealPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockEcarteGame) GetMatchScore(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockEcarteGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetPlayer(i int) *domain.EcartePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.EcartePlayer)
	}
	return nil
}

func (m *MockEcarteGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockEcarteGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockEcarteGame) GetHint() *domain.EcarteHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.EcarteHint)
	}
	return nil
}

func (m *MockEcarteGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
