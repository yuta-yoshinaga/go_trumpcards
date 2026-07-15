//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDeuceToSevenGame is the testify/mock implementation of DeuceToSevenGame.
type MockDeuceToSevenGame struct {
	mock.Mock
}

// Reset mock.
func (m *MockDeuceToSevenGame) Reset() error {
	ret := m.Called()
	return ret.Error(0)
}

// PlayerAction mock.
func (m *MockDeuceToSevenGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// PlayerExchange mock.
func (m *MockDeuceToSevenGame) PlayerExchange(indices []int) error {
	ret := m.Called(indices)
	return ret.Error(0)
}

// PlayerStand mock.
func (m *MockDeuceToSevenGame) PlayerStand() error {
	ret := m.Called()
	return ret.Error(0)
}

// GetPlayers mock.
func (m *MockDeuceToSevenGame) GetPlayers() []*domain.DeuceToSevenPlayer {
	ret := m.Called()
	return ret.Get(0).([]*domain.DeuceToSevenPlayer)
}

// SuggestExchange mock.
func (m *MockDeuceToSevenGame) SuggestExchange(playerIdx int) []int {
	ret := m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetPhase mock.
func (m *MockDeuceToSevenGame) GetPhase() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetDrawIndex mock.
func (m *MockDeuceToSevenGame) GetDrawIndex() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetPot mock.
func (m *MockDeuceToSevenGame) GetPot() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetSidePots mock.
func (m *MockDeuceToSevenGame) GetSidePots() []domain.SidePot {
	ret := m.Called()
	return ret.Get(0).([]domain.SidePot)
}

// GetDealerIdx mock.
func (m *MockDeuceToSevenGame) GetDealerIdx() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn mock.
func (m *MockDeuceToSevenGame) GetCurrentTurn() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetGameEndFlag mock.
func (m *MockDeuceToSevenGame) GetGameEndFlag() bool {
	ret := m.Called()
	return ret.Get(0).(bool)
}

// GetLastBet mock.
func (m *MockDeuceToSevenGame) GetLastBet() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetMinRaise mock.
func (m *MockDeuceToSevenGame) GetMinRaise() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount mock.
func (m *MockDeuceToSevenGame) GetRaiseCount() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetAnte mock.
func (m *MockDeuceToSevenGame) GetAnte() int {
	ret := m.Called()
	return ret.Get(0).(int)
}

// GetRoundResults mock.
func (m *MockDeuceToSevenGame) GetRoundResults() []domain.DeuceToSevenResult {
	ret := m.Called()
	return ret.Get(0).([]domain.DeuceToSevenResult)
}

// GetCpuActions mock.
func (m *MockDeuceToSevenGame) GetCpuActions() []domain.DeuceToSevenCpuAction {
	ret := m.Called()
	return ret.Get(0).([]domain.DeuceToSevenCpuAction)
}

// GetCpuExchanges mock.
func (m *MockDeuceToSevenGame) GetCpuExchanges() []domain.DeuceToSevenCpuExchange {
	ret := m.Called()
	return ret.Get(0).([]domain.DeuceToSevenCpuExchange)
}

// GetConfig mock.
func (m *MockDeuceToSevenGame) GetConfig() domain.DeuceToSevenConfig {
	ret := m.Called()
	return ret.Get(0).(domain.DeuceToSevenConfig)
}

// SetConfig mock.
func (m *MockDeuceToSevenGame) SetConfig(cfg domain.DeuceToSevenConfig) {
	m.Called(cfg)
}

// GetHumanProfile mock.
func (m *MockDeuceToSevenGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := m.Called()
	val := ret.Get(0)
	if val == nil {
		return nil
	}
	return val.(*domain.BettingHumanProfile)
}

// ResetProfile mock.
func (m *MockDeuceToSevenGame) ResetProfile() { m.Called() }

// ExportProfile mock.
func (m *MockDeuceToSevenGame) ExportProfile() any {
	ret := m.Called()
	return ret.Get(0)
}

// ImportProfile mock.
func (m *MockDeuceToSevenGame) ImportProfile(data []byte) error {
	ret := m.Called(data)
	return ret.Error(0)
}

// GetActionLog mock.
func (m *MockDeuceToSevenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := m.Called()
	return ret.Get(0).([]*domain.ActionLogEntry)
}
