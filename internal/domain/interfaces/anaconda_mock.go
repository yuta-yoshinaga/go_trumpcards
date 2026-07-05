//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAnacondaGame はアナコンダ (Anaconda) のゲームモック。
type MockAnacondaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockAnacondaGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockAnacondaGame) NextRound() { _m.Called() }

// Pass モック
func (_m *MockAnacondaGame) Pass(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// Keep モック
func (_m *MockAnacondaGame) Keep(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// PlayerCall モック
func (_m *MockAnacondaGame) PlayerCall() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockAnacondaGame) PlayerRaise() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockAnacondaGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockAnacondaGame) GetConfig() domain.AnacondaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.AnacondaConfig)
}

// SetConfig モック
func (_m *MockAnacondaGame) SetConfig(cfg domain.AnacondaConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockAnacondaGame) GetPhase() domain.AnacondaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AnacondaPhase)
}

// GetGameEndFlag モック
func (_m *MockAnacondaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockAnacondaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockAnacondaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockAnacondaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPassCount モック
func (_m *MockAnacondaGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRollIndex モック
func (_m *MockAnacondaGame) GetRollIndex() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockAnacondaGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBet モック
func (_m *MockAnacondaGame) GetCurrentBet() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount モック
func (_m *MockAnacondaGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMaxRaises モック
func (_m *MockAnacondaGame) GetMaxRaises() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockAnacondaGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockAnacondaGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockAnacondaGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockAnacondaGame) GetResult() domain.AnacondaResult {
	ret := _m.Called()
	return ret.Get(0).(domain.AnacondaResult)
}

// GetPlayerCnt モック
func (_m *MockAnacondaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockAnacondaGame) GetPlayer(i int) *domain.AnacondaPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.AnacondaPlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockAnacondaGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRevealedCards モック
func (_m *MockAnacondaGame) GetRevealedCards(idx int) []*domain.Card {
	ret := _m.Called(idx)
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// IsHandFullyRevealed モック
func (_m *MockAnacondaGame) IsHandFullyRevealed(idx int) bool {
	ret := _m.Called(idx)
	return ret.Get(0).(bool)
}

// IsHumanTurn モック
func (_m *MockAnacondaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanRaise モック
func (_m *MockAnacondaGame) CanRaise() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHint モック
func (_m *MockAnacondaGame) GetHint() *domain.AnacondaHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.AnacondaHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockAnacondaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
