//go:build test && (!js || !wasm || extra4)

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPutGame プットゲームモック
type MockPutGame struct {
	mock.Mock
}

func (m *MockPutGame) Reset() {
	m.Called()
}

func (m *MockPutGame) PlayerPlay(cardIndex int) error {
	args := m.Called(cardIndex)
	return args.Error(0)
}

func (m *MockPutGame) DeclarePut() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPutGame) RespondPut(accept bool) error {
	args := m.Called(accept)
	return args.Error(0)
}

func (m *MockPutGame) CpuStep() {
	m.Called()
}

func (m *MockPutGame) Next() {
	m.Called()
}

func (m *MockPutGame) GetConfig() domain.PutConfig {
	args := m.Called()
	return args.Get(0).(domain.PutConfig)
}

func (m *MockPutGame) SetConfig(cfg domain.PutConfig) {
	m.Called(cfg)
}

func (m *MockPutGame) GetGameEndFlag() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPutGame) GetPhase() domain.PutPhase {
	args := m.Called()
	return args.Get(0).(domain.PutPhase)
}

func (m *MockPutGame) IsHumanTurn() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPutGame) CanDeclarePut() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPutGame) GetHandNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetTrickNumber() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetCurrentPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetResponderIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetCurrentTrick() []*domain.TrickCard {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

func (m *MockPutGame) GetTrickResults() []int {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockPutGame) GetLeadPlayerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetManoIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetDealerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetHandStake() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetAcceptedLevel() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetPendingLevel() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetPutCallerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetMatchTarget() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetPlayerMatchPoints(i int) int {
	args := m.Called(i)
	return args.Int(0)
}

func (m *MockPutGame) GetHandWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetWinnerIdx() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetPlayerCnt() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockPutGame) GetPlayer(i int) *domain.PutPlayer {
	args := m.Called(i)
	if v := args.Get(0); v != nil {
		return v.(*domain.PutPlayer)
	}
	return nil
}

func (m *MockPutGame) GetValidPlayIndices(playerIdx int) []int {
	args := m.Called(playerIdx)
	if v := args.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

func (m *MockPutGame) GetHint() *domain.PutHint {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.(*domain.PutHint)
	}
	return nil
}

func (m *MockPutGame) GetActionLog() []*domain.ActionLogEntry {
	args := m.Called()
	if v := args.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
