//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGermanWhistGame ジャーマンホイストゲームモック
type MockGermanWhistGame struct {
	mock.Mock
}

func (m *MockGermanWhistGame) Reset() {
	m.Called()
}

func (m *MockGermanWhistGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockGermanWhistGame) CpuPlay() {
	m.Called()
}

func (m *MockGermanWhistGame) GiveUp() {
	m.Called()
}

func (m *MockGermanWhistGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGermanWhistGame) GetPhase() domain.GermanWhistPhase {
	args := m.Called()
	return args.Get(0).(domain.GermanWhistPhase)
}

func (m *MockGermanWhistGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockGermanWhistGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.TrickCard)
}

func (m *MockGermanWhistGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetUpCard() *domain.Card {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.Card)
}

func (m *MockGermanWhistGame) GetStockCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]int)
}

func (m *MockGermanWhistGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetPlayer(i int) *domain.GermanWhistPlayer {
	args := m.Called(i)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.GermanWhistPlayer)
}

func (m *MockGermanWhistGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockGermanWhistGame) GetHint() *domain.GermanWhistHint {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.GermanWhistHint)
}

func (m *MockGermanWhistGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*domain.ActionLogEntry)
}
