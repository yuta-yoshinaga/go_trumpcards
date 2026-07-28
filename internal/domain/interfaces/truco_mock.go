//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrucoGame トゥルコゲームモック
type MockTrucoGame struct {
	mock.Mock
}

func (m *MockTrucoGame) Reset() {
	m.Called()
}

func (m *MockTrucoGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockTrucoGame) DeclareTruco() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTrucoGame) RespondTruco(accept bool) error {
	args := m.Called(accept)
	return args.Error(0)
}

func (m *MockTrucoGame) CpuStep() {
	m.Called()
}

func (m *MockTrucoGame) Next() {
	m.Called()
}

func (m *MockTrucoGame) GetConfig() domain.TrucoConfig {
	args := m.Called()
	return args.Get(0).(domain.TrucoConfig)
}

func (m *MockTrucoGame) SetConfig(cfg domain.TrucoConfig) {
	m.Called(cfg)
}

func (m *MockTrucoGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTrucoGame) GetPhase() domain.TrucoPhase {
	args := m.Called()
	return args.Get(0).(domain.TrucoPhase)
}

func (m *MockTrucoGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTrucoGame) CanDeclareTruco() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockTrucoGame) GetHandNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetResponderIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockTrucoGame) GetTrickResults() []int {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockTrucoGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetManoIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetHandStake() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetAcceptedLevel() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetPendingLevel() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetTrucoCallerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetMatchTarget() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetPlayerMatchPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockTrucoGame) GetHandWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockTrucoGame) GetPlayer(i int) *domain.TrucoPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.TrucoPlayer)
	}
	return nil
}

func (m *MockTrucoGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockTrucoGame) GetHint() *domain.TrucoHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.TrucoHint)
	}
	return nil
}

func (m *MockTrucoGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
