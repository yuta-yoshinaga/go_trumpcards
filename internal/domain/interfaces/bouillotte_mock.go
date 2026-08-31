//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBouillotteGame はブイヨット (Bouillotte) のゲームモック。
type MockBouillotteGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockBouillotteGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockBouillotteGame) NextRound() { _m.Called() }

// PlayerCall モック
func (_m *MockBouillotteGame) PlayerCall() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockBouillotteGame) PlayerRaise() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockBouillotteGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockBouillotteGame) GetConfig() domain.BouillotteConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.BouillotteConfig)
}

// SetConfig モック
func (_m *MockBouillotteGame) SetConfig(cfg domain.BouillotteConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockBouillotteGame) GetPhase() domain.BouillottePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BouillottePhase)
}

// GetGameEndFlag モック
func (_m *MockBouillotteGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockBouillotteGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockBouillotteGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockBouillotteGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockBouillotteGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBet モック
func (_m *MockBouillotteGame) GetCurrentBet() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRaiseCount モック
func (_m *MockBouillotteGame) GetRaiseCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMaxRaises モック
func (_m *MockBouillotteGame) GetMaxRaises() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockBouillotteGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRetourne モック
func (_m *MockBouillotteGame) GetRetourne() *domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.Card)
	}
	return nil
}

// GetWinnerIdx モック
func (_m *MockBouillotteGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockBouillotteGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockBouillotteGame) GetResult() domain.BouillotteResult {
	ret := _m.Called()
	return ret.Get(0).(domain.BouillotteResult)
}

// GetPlayerCnt モック
func (_m *MockBouillotteGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockBouillotteGame) GetPlayer(i int) *domain.BouillottePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.BouillottePlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockBouillotteGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsHumanTurn モック
func (_m *MockBouillotteGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanRaise モック
func (_m *MockBouillotteGame) CanRaise() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHint モック
func (_m *MockBouillotteGame) GetHint() *domain.BouillotteHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.BouillotteHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockBouillotteGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}

// AnalyzeRetourneMatch モック
func (_m *MockBouillotteGame) AnalyzeRetourneMatch(playerIdx int) *domain.BouillotteRetourneMatch {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(*domain.BouillotteRetourneMatch)
	}
	return nil
}
