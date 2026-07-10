//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPrimeroGame はプリメロ (Primero) のゲームモック。
type MockPrimeroGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPrimeroGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockPrimeroGame) NextRound() { _m.Called() }

// PlayerCall モック
func (_m *MockPrimeroGame) PlayerCall() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockPrimeroGame) PlayerRaise() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockPrimeroGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockPrimeroGame) GetConfig() domain.PrimeroConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PrimeroConfig)
}

// SetConfig モック
func (_m *MockPrimeroGame) SetConfig(cfg domain.PrimeroConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockPrimeroGame) GetPhase() domain.PrimeroPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PrimeroPhase)
}

// GetGameEndFlag モック
func (_m *MockPrimeroGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockPrimeroGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockPrimeroGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockPrimeroGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockPrimeroGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBet モック
func (_m *MockPrimeroGame) GetCurrentBet() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount モック
func (_m *MockPrimeroGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMaxRaises モック
func (_m *MockPrimeroGame) GetMaxRaises() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockPrimeroGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockPrimeroGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockPrimeroGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockPrimeroGame) GetResult() domain.PrimeroResult {
	ret := _m.Called()
	return ret.Get(0).(domain.PrimeroResult)
}

// GetPlayerCnt モック
func (_m *MockPrimeroGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockPrimeroGame) GetPlayer(i int) *domain.PrimeroPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.PrimeroPlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockPrimeroGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsHumanTurn モック
func (_m *MockPrimeroGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanRaise モック
func (_m *MockPrimeroGame) CanRaise() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHint モック
func (_m *MockPrimeroGame) GetHint() *domain.PrimeroHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.PrimeroHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockPrimeroGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
