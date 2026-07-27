//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOmbreGame オンブル (Ombre) のゲームモック
type MockOmbreGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockOmbreGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockOmbreGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockOmbreGame) PlayerBid(bid domain.OmbreBid, trumpSuit int) error {
	ret := _m.Called(bid, trumpSuit)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockOmbreGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockOmbreGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockOmbreGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockOmbreGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockOmbreGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockOmbreGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockOmbreGame) GetConfig() domain.OmbreConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbreConfig)
}

// SetConfig モック
func (_m *MockOmbreGame) SetConfig(cfg domain.OmbreConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockOmbreGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockOmbreGame) GetPhase() domain.OmbrePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbrePhase)
}

// IsHumanTurn モック
func (_m *MockOmbreGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockOmbreGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockOmbreGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockOmbreGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockOmbreGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockOmbreGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockOmbreGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockOmbreGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetForehandIdx モック
func (_m *MockOmbreGame) GetForehandIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetOmbreIdx モック
func (_m *MockOmbreGame) GetOmbreIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinningBid モック
func (_m *MockOmbreGame) GetWinningBid() domain.OmbreBid {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbreBid)
}

// GetTrumpSuit モック
func (_m *MockOmbreGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBidderIdx モック
func (_m *MockOmbreGame) GetCurrentBidderIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockOmbreGame) GetPlayerScores() [domain.OmbrePlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.OmbrePlayerCnt]int)
}

// GetOutcome モック
func (_m *MockOmbreGame) GetOutcome() domain.OmbreOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbreOutcome)
}

// GetResult モック
func (_m *MockOmbreGame) GetResult() domain.OmbreResult {
	ret := _m.Called()
	return ret.Get(0).(domain.OmbreResult)
}

// GetWinnerPlayer モック
func (_m *MockOmbreGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockOmbreGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockOmbreGame) GetPlayer(i int) *domain.OmbrePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.OmbrePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockOmbreGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockOmbreGame) GetHint() *domain.OmbreHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.OmbreHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockOmbreGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
