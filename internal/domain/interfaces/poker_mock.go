//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPokerGame ポーカーゲームモック
type MockPokerGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerGame) Reset() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerAction モック
func (_m *MockPokerGame) PlayerAction(action, amount, humanPlayMs int) error {
	ret := _m.Called(action, amount, humanPlayMs)
	return ret.Error(0)
}

// PlayerExchange モック
func (_m *MockPokerGame) PlayerExchange(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// PlayerStand モック
func (_m *MockPokerGame) PlayerStand() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayers モック
func (_m *MockPokerGame) GetPlayers() []*domain.PokerPlayer {
	ret := _m.Called()
	return ret.Get(0).([]*domain.PokerPlayer)
}

// GetPhase モック
func (_m *MockPokerGame) GetPhase() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsExchangeRead モック
func (_m *MockPokerGame) IsExchangeRead(playerIdx int) bool {
	ret := _m.Called(playerIdx)
	return ret.Bool(0)
}

// GetPot モック
// GetEquity はエクイティを返すモック。
func (_m *MockPokerGame) GetEquity() *domain.HoldemEquityResult {
	out, _ := _m.Called().Get(0).(*domain.HoldemEquityResult)
	return out
}

// GetPotOdds はポットオッズを返すモック。
func (_m *MockPokerGame) GetPotOdds() float64 {
	return _m.Called().Get(0).(float64)
}

func (_m *MockPokerGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSidePots モック
func (_m *MockPokerGame) GetSidePots() []domain.SidePot {
	ret := _m.Called()
	return ret.Get(0).([]domain.SidePot)
}

// GetDealerIdx モック
func (_m *MockPokerGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTurn モック
func (_m *MockPokerGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetGameEndFlag モック
func (_m *MockPokerGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetLastBet モック
func (_m *MockPokerGame) GetLastBet() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMinRaise モック
func (_m *MockPokerGame) GetMinRaise() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount モック
func (_m *MockPokerGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockPokerGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundResults モック
func (_m *MockPokerGame) GetRoundResults() []domain.PokerResult {
	ret := _m.Called()
	return ret.Get(0).([]domain.PokerResult)
}

// GetCpuActions モック
func (_m *MockPokerGame) GetCpuActions() []domain.PokerCpuAction {
	ret := _m.Called()
	return ret.Get(0).([]domain.PokerCpuAction)
}

// GetCpuExchanges モック
func (_m *MockPokerGame) GetCpuExchanges() []domain.PokerCpuExchange {
	ret := _m.Called()
	return ret.Get(0).([]domain.PokerCpuExchange)
}

// GetConfig モック
func (_m *MockPokerGame) GetConfig() domain.PokerConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PokerConfig)
}

// SetConfig モック
func (_m *MockPokerGame) SetConfig(cfg domain.PokerConfig) {
	_m.Called(cfg)
}

// GetHumanProfile モック
func (_m *MockPokerGame) GetHumanProfile() *domain.BettingHumanProfile {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.BettingHumanProfile); ok {
		return val
	}
	return nil
}

// ResetProfile モック
func (_m *MockPokerGame) ResetProfile() {
	_m.Called()
}

// ExportProfile モック
func (_m *MockPokerGame) ExportProfile() interface{} {
	ret := _m.Called()
	return ret.Get(0)
}

// ImportProfile モック
func (_m *MockPokerGame) ImportProfile(data []byte) error {
	ret := _m.Called(data)
	return ret.Error(0)
}

// GetActionLog モック
func (_m *MockPokerGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	return ret.Get(0).([]*domain.ActionLogEntry)
}

// CalcDrawOdds モック
func (_m *MockPokerGame) CalcDrawOdds(indices []int) ([]domain.PokerDrawOdds, error) {
	ret := _m.Called(indices)
	return ret.Get(0).([]domain.PokerDrawOdds), ret.Error(1)
}
