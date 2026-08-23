//go:build test && (!js || !wasm || classic)

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBrusquembilleGame ブリュスカンビーユゲームモック
type MockBrusquembilleGame struct {
	mock.Mock
}

func (m *MockBrusquembilleGame) Reset() {
	m.Called()
}

func (m *MockBrusquembilleGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockBrusquembilleGame) CpuPlay() {
	m.Called()
}

func (m *MockBrusquembilleGame) ResolveTrick() {
	m.Called()
}

func (m *MockBrusquembilleGame) NextTrick() {
	m.Called()
}

func (m *MockBrusquembilleGame) GetConfig() domain.BrusquembilleConfig {
	args := m.Called()
	return args.Get(0).(domain.BrusquembilleConfig)
}

func (m *MockBrusquembilleGame) SetConfig(cfg domain.BrusquembilleConfig) {
	m.Called(cfg)
}

func (m *MockBrusquembilleGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBrusquembilleGame) IsFollowRequired() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBrusquembilleGame) GetPhase() domain.BrusquembillePhase {
	args := m.Called()
	return args.Get(0).(domain.BrusquembillePhase)
}

func (m *MockBrusquembilleGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBrusquembilleGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockBrusquembilleGame) GetTrumpSuit() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetTrumpCard() *domain.Card {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

func (m *MockBrusquembilleGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetPlayerPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetPlayer(i int) *domain.BrusquembillePlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.BrusquembillePlayer)
	}
	return nil
}

func (m *MockBrusquembilleGame) GetStockRemaining() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockBrusquembilleGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockBrusquembilleGame) GetHint() *domain.BrusquembilleHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.BrusquembilleHint)
	}
	return nil
}

func (m *MockBrusquembilleGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
