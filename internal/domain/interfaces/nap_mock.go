//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNapGame ナップのゲームモック
type MockNapGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockNapGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockNapGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockNapGame) PlayerBid(bid domain.NapBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockNapGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockNapGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockNapGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockNapGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockNapGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockNapGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockNapGame) GetConfig() domain.NapConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.NapConfig)
}

// SetConfig モック
func (_m *MockNapGame) SetConfig(cfg domain.NapConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockNapGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockNapGame) GetPhase() domain.NapPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.NapPhase)
}

// IsHumanTurn モック
func (_m *MockNapGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockNapGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockNapGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockNapGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockNapGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockNapGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockNapGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockNapGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockNapGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockNapGame) GetContract() domain.NapBid {
	ret := _m.Called()
	return ret.Get(0).(domain.NapBid)
}

// GetBids モック
func (_m *MockNapGame) GetBids() [domain.NapPlayerCnt]domain.NapBid {
	ret := _m.Called()
	return ret.Get(0).([domain.NapPlayerCnt]domain.NapBid)
}

// GetTrumpSuit モック
func (_m *MockNapGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockNapGame) GetPlayerScores() [domain.NapPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.NapPlayerCnt]int)
}

// GetRoundTricks モック
func (_m *MockNapGame) GetRoundTricks() [domain.NapPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.NapPlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockNapGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockNapGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockNapGame) GetPlayer(i int) *domain.NapPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.NapPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockNapGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockNapGame) GetHint() *domain.NapHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.NapHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockNapGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
