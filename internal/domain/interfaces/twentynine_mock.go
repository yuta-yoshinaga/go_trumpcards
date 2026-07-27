//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTwentyNineGame トゥエンティナイン (29) のゲームモック
type MockTwentyNineGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTwentyNineGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTwentyNineGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockTwentyNineGame) PlayerBid(bid domain.TwentyNineBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockTwentyNineGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockTwentyNineGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTwentyNineGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTwentyNineGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTwentyNineGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTwentyNineGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTwentyNineGame) GetConfig() domain.TwentyNineConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TwentyNineConfig)
}

// SetConfig モック
func (_m *MockTwentyNineGame) SetConfig(cfg domain.TwentyNineConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTwentyNineGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTwentyNineGame) GetPhase() domain.TwentyNinePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TwentyNinePhase)
}

// IsHumanTurn モック
func (_m *MockTwentyNineGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockTwentyNineGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockTwentyNineGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockTwentyNineGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTwentyNineGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockTwentyNineGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockTwentyNineGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockTwentyNineGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockTwentyNineGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockTwentyNineGame) GetContract() domain.TwentyNineBid {
	ret := _m.Called()
	return ret.Get(0).(domain.TwentyNineBid)
}

// GetBids モック
func (_m *MockTwentyNineGame) GetBids() [domain.TwentyNinePlayerCnt]domain.TwentyNineBid {
	ret := _m.Called()
	return ret.Get(0).([domain.TwentyNinePlayerCnt]domain.TwentyNineBid)
}

// GetTrumpSuit モック
func (_m *MockTwentyNineGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpRevealed モック
func (_m *MockTwentyNineGame) GetTrumpRevealed() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetTeamScores モック
func (_m *MockTwentyNineGame) GetTeamScores() [domain.TwentyNineTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TwentyNineTeamCnt]int)
}

// GetRoundTeamPoints モック
func (_m *MockTwentyNineGame) GetRoundTeamPoints() [domain.TwentyNineTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TwentyNineTeamCnt]int)
}

// GetWinnerTeam モック
func (_m *MockTwentyNineGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTwentyNineGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTwentyNineGame) GetPlayer(i int) *domain.TwentyNinePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TwentyNinePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTwentyNineGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTwentyNineGame) GetHint() *domain.TwentyNineHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TwentyNineHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTwentyNineGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
