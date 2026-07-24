//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSoloWhistGame ソロ・ホイストのゲームモック
type MockSoloWhistGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSoloWhistGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSoloWhistGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockSoloWhistGame) PlayerBid(bid domain.SoloWhistBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockSoloWhistGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockSoloWhistGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSoloWhistGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSoloWhistGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSoloWhistGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSoloWhistGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSoloWhistGame) GetConfig() domain.SoloWhistConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SoloWhistConfig)
}

// SetConfig モック
func (_m *MockSoloWhistGame) SetConfig(cfg domain.SoloWhistConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSoloWhistGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSoloWhistGame) GetPhase() domain.SoloWhistPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SoloWhistPhase)
}

// IsHumanTurn モック
func (_m *MockSoloWhistGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockSoloWhistGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSoloWhistGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSoloWhistGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSoloWhistGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSoloWhistGame) GetCurrentTrick() []*domain.SoloWhistTrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.SoloWhistTrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSoloWhistGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSoloWhistGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockSoloWhistGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockSoloWhistGame) GetContract() domain.SoloWhistBid {
	ret := _m.Called()
	return ret.Get(0).(domain.SoloWhistBid)
}

// GetBids モック
func (_m *MockSoloWhistGame) GetBids() [domain.SoloWhistPlayerCnt]domain.SoloWhistBid {
	ret := _m.Called()
	return ret.Get(0).([domain.SoloWhistPlayerCnt]domain.SoloWhistBid)
}

// GetBidDone モック
func (_m *MockSoloWhistGame) GetBidDone() [domain.SoloWhistPlayerCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.SoloWhistPlayerCnt]bool)
}

// GetTrumpSuit モック
func (_m *MockSoloWhistGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockSoloWhistGame) GetPlayerScores() [domain.SoloWhistPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SoloWhistPlayerCnt]int)
}

// GetRoundTricks モック
func (_m *MockSoloWhistGame) GetRoundTricks() [domain.SoloWhistPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.SoloWhistPlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockSoloWhistGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSoloWhistGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSoloWhistGame) GetPlayer(i int) *domain.SoloWhistPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SoloWhistPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSoloWhistGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSoloWhistGame) GetHint() *domain.SoloWhistHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SoloWhistHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSoloWhistGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
