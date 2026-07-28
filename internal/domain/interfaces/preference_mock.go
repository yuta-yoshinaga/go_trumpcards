//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPreferenceGame プレフェランスのゲームモック
type MockPreferenceGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPreferenceGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockPreferenceGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockPreferenceGame) PlayerBid(bid domain.PreferenceBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockPreferenceGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockPreferenceGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockPreferenceGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockPreferenceGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockPreferenceGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockPreferenceGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockPreferenceGame) GetConfig() domain.PreferenceConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.PreferenceConfig)
}

// SetConfig モック
func (_m *MockPreferenceGame) SetConfig(cfg domain.PreferenceConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockPreferenceGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockPreferenceGame) GetPhase() domain.PreferencePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PreferencePhase)
}

// IsHumanTurn モック
func (_m *MockPreferenceGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockPreferenceGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockPreferenceGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockPreferenceGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockPreferenceGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockPreferenceGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockPreferenceGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockPreferenceGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockPreferenceGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockPreferenceGame) GetContract() domain.PreferenceBid {
	ret := _m.Called()
	return ret.Get(0).(domain.PreferenceBid)
}

// GetBids モック
func (_m *MockPreferenceGame) GetBids() [domain.PreferencePlayerCnt]domain.PreferenceBid {
	ret := _m.Called()
	return ret.Get(0).([domain.PreferencePlayerCnt]domain.PreferenceBid)
}

// GetTrumpSuit モック
func (_m *MockPreferenceGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockPreferenceGame) GetPlayerScores() [domain.PreferencePlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.PreferencePlayerCnt]int)
}

// GetRoundTricks モック
func (_m *MockPreferenceGame) GetRoundTricks() [domain.PreferencePlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.PreferencePlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockPreferenceGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockPreferenceGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockPreferenceGame) GetPlayer(i int) *domain.PreferencePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.PreferencePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockPreferenceGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockPreferenceGame) GetHint() *domain.PreferenceHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.PreferenceHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockPreferenceGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
