//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKempsGame はケムプスのゲームモック。
type MockKempsGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKempsGame) Reset() { _m.Called() }

// ResetWithConfig モック
func (_m *MockKempsGame) ResetWithConfig(cfg domain.KempsConfig) { _m.Called(cfg) }

// NextRound モック
func (_m *MockKempsGame) NextRound() { _m.Called() }

// PlayerSwap モック
func (_m *MockKempsGame) PlayerSwap(handIndex, fieldIndex int) error {
	ret := _m.Called(handIndex, fieldIndex)
	return ret.Error(0)
}

// PlayerPass モック
func (_m *MockKempsGame) PlayerPass() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerSetSignal モック
func (_m *MockKempsGame) PlayerSetSignal(signalType int) { _m.Called(signalType) }

// PlayerDeclareKemps モック
func (_m *MockKempsGame) PlayerDeclareKemps() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerDeclareCounterKemps モック
func (_m *MockKempsGame) PlayerDeclareCounterKemps(targetSeat int) error {
	ret := _m.Called(targetSeat)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockKempsGame) CpuPlay() { _m.Called() }

// GetConfig モック
func (_m *MockKempsGame) GetConfig() domain.KempsConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KempsConfig)
}

// SetConfig モック
func (_m *MockKempsGame) SetConfig(cfg domain.KempsConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockKempsGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockKempsGame) GetPhase() domain.KempsPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KempsPhase)
}

// IsHumanTurn モック
func (_m *MockKempsGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPlayerCnt モック
func (_m *MockKempsGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockKempsGame) GetPlayer(i int) *domain.KempsPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KempsPlayer)
	}
	return nil
}

// GetWinnerTeam モック
func (_m *MockKempsGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScore モック
func (_m *MockKempsGame) GetTeamScore(team int) int {
	ret := _m.Called(team)
	return ret.Get(0).(int)
}

// GetFieldSize モック
func (_m *MockKempsGame) GetFieldSize() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFieldCard モック
func (_m *MockKempsGame) GetFieldCard(i int) *domain.Card {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetCurrentPlayerIdx モック
func (_m *MockKempsGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSignalType モック
func (_m *MockKempsGame) GetSignalType() domain.SignalType {
	ret := _m.Called()
	return ret.Get(0).(domain.SignalType)
}

// GetFourHolderIdx モック
func (_m *MockKempsGame) GetFourHolderIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsPartnerSignaling モック
func (_m *MockKempsGame) IsPartnerSignaling() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsOpponentSignaling モック
func (_m *MockKempsGame) IsOpponentSignaling() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundResult モック
func (_m *MockKempsGame) GetRoundResult() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinnerTeam モック
func (_m *MockKempsGame) GetRoundWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockKempsGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetActionLog モック
func (_m *MockKempsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
