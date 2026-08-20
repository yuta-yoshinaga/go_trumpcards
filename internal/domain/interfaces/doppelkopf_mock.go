//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoppelkopfGame ドッペルコップのゲームモック
type MockDoppelkopfGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoppelkopfGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockDoppelkopfGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockDoppelkopfGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerAnnounce モック
func (_m *MockDoppelkopfGame) PlayerAnnounce() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockDoppelkopfGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockDoppelkopfGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockDoppelkopfGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockDoppelkopfGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockDoppelkopfGame) GetConfig() domain.DoppelkopfConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.DoppelkopfConfig)
}

// SetConfig モック
func (_m *MockDoppelkopfGame) SetConfig(cfg domain.DoppelkopfConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockDoppelkopfGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockDoppelkopfGame) GetPhase() domain.DoppelkopfPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.DoppelkopfPhase)
}

// IsHumanTurn モック
func (_m *MockDoppelkopfGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockDoppelkopfGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockDoppelkopfGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockDoppelkopfGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockDoppelkopfGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockDoppelkopfGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockDoppelkopfGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsRe モック
func (_m *MockDoppelkopfGame) IsRe(playerIdx int) bool {
	ret := _m.Called(playerIdx)
	return ret.Get(0).(bool)
}

// IsSoloRe モック
func (_m *MockDoppelkopfGame) IsSoloRe() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// AreTeamsRevealed モック
func (_m *MockDoppelkopfGame) AreTeamsRevealed() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsReAnnounced モック
func (_m *MockDoppelkopfGame) IsReAnnounced() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsKontraAnnounced モック
func (_m *MockDoppelkopfGame) IsKontraAnnounced() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanHumanAnnounce モック
func (_m *MockDoppelkopfGame) CanHumanAnnounce() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundRePoints モック
func (_m *MockDoppelkopfGame) GetRoundRePoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundReWon モック
func (_m *MockDoppelkopfGame) GetRoundReWon() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundGamePoints モック
func (_m *MockDoppelkopfGame) GetRoundGamePoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockDoppelkopfGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockDoppelkopfGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockDoppelkopfGame) GetPlayer(i int) *domain.DoppelkopfPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.DoppelkopfPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockDoppelkopfGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetTrumpIndices モック
func (_m *MockDoppelkopfGame) GetTrumpIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]int)
}

// GetHint モック
func (_m *MockDoppelkopfGame) GetHint() *domain.DoppelkopfHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.DoppelkopfHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockDoppelkopfGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
