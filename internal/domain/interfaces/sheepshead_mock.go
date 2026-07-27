//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSheepsheadGame シープスヘッドのゲームモック
type MockSheepsheadGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSheepsheadGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSheepsheadGame) NextRound() {
	_m.Called()
}

// PlayerPick モック
func (_m *MockSheepsheadGame) PlayerPick(pick bool) error {
	ret := _m.Called(pick)
	return ret.Error(0)
}

// PlayerBury モック
func (_m *MockSheepsheadGame) PlayerBury(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// PlayerCall モック
func (_m *MockSheepsheadGame) PlayerCall(suit int) error {
	ret := _m.Called(suit)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockSheepsheadGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSheepsheadGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSheepsheadGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSheepsheadGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSheepsheadGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSheepsheadGame) GetConfig() domain.SheepsheadConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SheepsheadConfig)
}

// SetConfig モック
func (_m *MockSheepsheadGame) SetConfig(cfg domain.SheepsheadConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSheepsheadGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSheepsheadGame) GetPhase() domain.SheepsheadPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SheepsheadPhase)
}

// IsHumanTurn モック
func (_m *MockSheepsheadGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSheepsheadGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSheepsheadGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSheepsheadGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSheepsheadGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSheepsheadGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSheepsheadGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBlind モック
func (_m *MockSheepsheadGame) GetBlind() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetBuried モック
func (_m *MockSheepsheadGame) GetBuried() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetPickerIdx モック
func (_m *MockSheepsheadGame) GetPickerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPartnerIdx モック
func (_m *MockSheepsheadGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCalledSuit モック
func (_m *MockSheepsheadGame) GetCalledSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsPartnerRevealed モック
func (_m *MockSheepsheadGame) IsPartnerRevealed() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPassCount モック
func (_m *MockSheepsheadGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundPickerPoints モック
func (_m *MockSheepsheadGame) GetRoundPickerPoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundMultiplier モック
func (_m *MockSheepsheadGame) GetRoundMultiplier() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundPickerWon モック
func (_m *MockSheepsheadGame) GetRoundPickerWon() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetWinnerIdx モック
func (_m *MockSheepsheadGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSheepsheadGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSheepsheadGame) GetPlayer(i int) *domain.SheepsheadPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SheepsheadPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSheepsheadGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCallableSuits モック
func (_m *MockSheepsheadGame) GetCallableSuits() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSheepsheadGame) GetHint() *domain.SheepsheadHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SheepsheadHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSheepsheadGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
