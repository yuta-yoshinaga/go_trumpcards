//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSchafkopfGame シャーフコップのゲームモック
type MockSchafkopfGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSchafkopfGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSchafkopfGame) NextRound() {
	_m.Called()
}

// PlayerDeclare モック
func (_m *MockSchafkopfGame) PlayerDeclare(pick bool, contract domain.SchafkopfContract, soloSuit int) error {
	ret := _m.Called(pick, contract, soloSuit)
	return ret.Error(0)
}

// GetContract モック
func (_m *MockSchafkopfGame) GetContract() domain.SchafkopfContract {
	ret := _m.Called()
	return ret.Get(0).(domain.SchafkopfContract)
}

// GetSoloSuit モック
func (_m *MockSchafkopfGame) GetSoloSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBeatableContracts モック
func (_m *MockSchafkopfGame) GetBeatableContracts() []domain.SchafkopfContract {
	ret := _m.Called()
	return ret.Get(0).([]domain.SchafkopfContract)
}

// PlayerCall モック
func (_m *MockSchafkopfGame) PlayerCall(suit int) error {
	ret := _m.Called(suit)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockSchafkopfGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSchafkopfGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSchafkopfGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSchafkopfGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSchafkopfGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSchafkopfGame) GetConfig() domain.SchafkopfConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SchafkopfConfig)
}

// SetConfig モック
func (_m *MockSchafkopfGame) SetConfig(cfg domain.SchafkopfConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSchafkopfGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSchafkopfGame) GetPhase() domain.SchafkopfPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SchafkopfPhase)
}

// IsHumanTurn モック
func (_m *MockSchafkopfGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSchafkopfGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSchafkopfGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSchafkopfGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSchafkopfGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSchafkopfGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSchafkopfGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPickerIdx モック
func (_m *MockSchafkopfGame) GetPickerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPartnerIdx モック
func (_m *MockSchafkopfGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCalledSuit モック
func (_m *MockSchafkopfGame) GetCalledSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsPartnerRevealed モック
func (_m *MockSchafkopfGame) IsPartnerRevealed() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPassCount モック
func (_m *MockSchafkopfGame) GetPassCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundPickerPoints モック
func (_m *MockSchafkopfGame) GetRoundPickerPoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundMultiplier モック
func (_m *MockSchafkopfGame) GetRoundMultiplier() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundPickerWon モック
func (_m *MockSchafkopfGame) GetRoundPickerWon() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetWinnerIdx モック
func (_m *MockSchafkopfGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSchafkopfGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSchafkopfGame) GetPlayer(i int) *domain.SchafkopfPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SchafkopfPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSchafkopfGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCallableSuits モック
func (_m *MockSchafkopfGame) GetCallableSuits() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSchafkopfGame) GetHint() *domain.SchafkopfHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SchafkopfHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSchafkopfGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
