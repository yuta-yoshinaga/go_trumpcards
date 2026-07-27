//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSchnapsenGame シュナプセンゲームモック
type MockSchnapsenGame struct {
	mock.Mock
}

func (m *MockSchnapsenGame) Reset() {
	m.Called()
}

func (m *MockSchnapsenGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockSchnapsenGame) PlayerDeclareMarriage(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockSchnapsenGame) CpuPlay() {
	m.Called()
}

func (m *MockSchnapsenGame) ResolveTrick() {
	m.Called()
}

func (m *MockSchnapsenGame) NextTrick() {
	m.Called()
}

func (m *MockSchnapsenGame) GetConfig() domain.SchnapsenConfig {
	args := m.Called()
	return args.Get(0).(domain.SchnapsenConfig)
}

func (m *MockSchnapsenGame) SetConfig(cfg domain.SchnapsenConfig) {
	m.Called(cfg)
}

func (m *MockSchnapsenGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSchnapsenGame) GetPhase() domain.SchnapsenPhase {
	args := m.Called()
	return args.Get(0).(domain.SchnapsenPhase)
}

func (m *MockSchnapsenGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSchnapsenGame) IsEndgame() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSchnapsenGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockSchnapsenGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockSchnapsenGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetPlayerPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetPlayer(i int) *domain.SchnapsenPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.SchnapsenPlayer)
	}
	return nil
}

func (m *MockSchnapsenGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSchnapsenGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockSchnapsenGame) GetMarriageIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockSchnapsenGame) GetHint() *domain.SchnapsenHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.SchnapsenHint)
	}
	return nil
}

func (m *MockSchnapsenGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
