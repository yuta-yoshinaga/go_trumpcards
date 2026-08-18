//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMusGame ムスのゲームモック
type MockMusGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMusGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockMusGame) NextRound() {
	_m.Called()
}

// PlayerMus モック
func (_m *MockMusGame) PlayerMus(mus bool) error {
	ret := _m.Called(mus)
	return ret.Error(0)
}

// PlayerDiscard モック
func (_m *MockMusGame) PlayerDiscard(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// PlayerBet モック
func (_m *MockMusGame) PlayerBet(action, amount int) error {
	ret := _m.Called(action, amount)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockMusGame) CpuPlay() {
	_m.Called()
}

// Showdown モック
func (_m *MockMusGame) Showdown() {
	_m.Called()
}

// GetConfig モック
func (_m *MockMusGame) GetConfig() domain.MusConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MusConfig)
}

// SetConfig モック
func (_m *MockMusGame) SetConfig(cfg domain.MusConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockMusGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockMusGame) GetPhase() domain.MusPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MusPhase)
}

// IsHumanTurn モック
func (_m *MockMusGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockMusGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetManoIdx モック
func (_m *MockMusGame) GetManoIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMusTurn モック
func (_m *MockMusGame) GetMusTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDiscardTurn モック
func (_m *MockMusGame) GetDiscardTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBetTeam モック
func (_m *MockMusGame) GetBetTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPendingStake モック
func (_m *MockMusGame) GetPendingStake() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastBettorTeam モック
func (_m *MockMusGame) GetLastBettorTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAmarrakos モック
func (_m *MockMusGame) GetAmarrakos() [domain.MusTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.MusTeamCnt]int)
}

// GetResult モック
func (_m *MockMusGame) GetResult(ri int) domain.MusRoundResult {
	ret := _m.Called(ri)
	return ret.Get(0).(domain.MusRoundResult)
}

// GetMusCycle モック
func (_m *MockMusGame) GetMusCycle() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerTeam モック
func (_m *MockMusGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockMusGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockMusGame) GetPlayer(i int) *domain.MusPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MusPlayer)
	}
	return nil
}

// GetHandSummary モック
func (_m *MockMusGame) GetHandSummary(i int) *domain.MusHandSummary {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MusHandSummary)
	}
	return nil
}

// GetHint モック
func (_m *MockMusGame) GetHint() *domain.MusHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.MusHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockMusGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
