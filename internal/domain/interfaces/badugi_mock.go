//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBadugiGame is the testify/mock implementation of BadugiGame.
type MockBadugiGame struct {
	mock.Mock
}

// Reset mock.
func (m *MockBadugiGame) Reset() error {
	ret := m.Called()
	return ret.Error(0)
}

// PlayerAction mock.
func (m *MockBadugiGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// PlayerExchange mock.
func (m *MockBadugiGame) PlayerExchange(indices []int, humanPlayMs int) error {
	ret := m.Called(indices, humanPlayMs)
	return ret.Error(0)
}

// PlayerStand mock.
func (m *MockBadugiGame) PlayerStand(humanPlayMs int) error {
	ret := m.Called(humanPlayMs)
	return ret.Error(0)
}

// GetPlayers mock.
func (m *MockBadugiGame) GetPlayers() []*domain.BadugiPlayer {
	ret := m.Called()
	return ret.Get(0).([]*domain.BadugiPlayer)
}

// GetPhase mock.
func (m *MockBadugiGame) GetPhase() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetDrawIndex mock.
func (m *MockBadugiGame) GetDrawIndex() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetPot mock.
func (m *MockBadugiGame) GetPot() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetSidePots mock.
func (m *MockBadugiGame) GetSidePots() []domain.SidePot {
	ret := m.Called()
	return ret.Get(0).([]domain.SidePot)
}

// GetDealerIdx mock.
func (m *MockBadugiGame) GetDealerIdx() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn mock.
func (m *MockBadugiGame) GetCurrentTurn() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetGameEndFlag mock.
func (m *MockBadugiGame) GetGameEndFlag() bool {
	ret := m.Called()
	return ret.Get(0).(bool)
}

// GetLastBet mock.
func (m *MockBadugiGame) GetLastBet() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetMinRaise mock.
func (m *MockBadugiGame) GetMinRaise() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount mock.
func (m *MockBadugiGame) GetRaiseCount() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetAnte mock.
func (m *MockBadugiGame) GetAnte() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetRoundResults mock.
func (m *MockBadugiGame) GetRoundResults() []domain.BadugiResult {
	ret := m.Called()
	return ret.Get(0).([]domain.BadugiResult)
}

// GetCpuActions mock.
func (m *MockBadugiGame) GetCpuActions() []domain.BadugiCpuAction {
	ret := m.Called()
	return ret.Get(0).([]domain.BadugiCpuAction)
}

// GetCpuExchanges mock.
func (m *MockBadugiGame) GetCpuExchanges() []domain.BadugiCpuExchange {
	ret := m.Called()
	return ret.Get(0).([]domain.BadugiCpuExchange)
}

// GetConfig mock.
func (m *MockBadugiGame) GetConfig() domain.BadugiConfig {
	ret := m.Called()
	return ret.Get(0).(domain.BadugiConfig)
}

// SetConfig mock.
func (m *MockBadugiGame) SetConfig(cfg domain.BadugiConfig) {
	m.Called(cfg)
}

// GetHumanProfile mock.
func (m *MockBadugiGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := m.Called()
	val := ret.Get(0)
	if val == nil {
		return nil
	}
	return val.(*domain.BettingHumanProfile)
}

// ResetProfile mock.
func (m *MockBadugiGame) ResetProfile() { m.Called() }

// ExportProfile mock.
func (m *MockBadugiGame) ExportProfile() any {
	ret := m.Called()
	return ret.Get(0)
}

// ImportProfile mock.
func (m *MockBadugiGame) ImportProfile(data []byte) error {
	ret := m.Called(data)
	return ret.Error(0)
}

// GetActionLog mock.
func (m *MockBadugiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := m.Called()
	return ret.Get(0).([]*domain.ActionLogEntry)
}
