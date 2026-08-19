//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFortyFivesGame オークション・フォーティファイブズのゲームモック
type MockFortyFivesGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockFortyFivesGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockFortyFivesGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockFortyFivesGame) PlayerBid(bid domain.FortyFivesBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockFortyFivesGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockFortyFivesGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockFortyFivesGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockFortyFivesGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockFortyFivesGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockFortyFivesGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockFortyFivesGame) GetConfig() domain.FortyFivesConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.FortyFivesConfig)
}

// SetConfig モック
func (_m *MockFortyFivesGame) SetConfig(cfg domain.FortyFivesConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockFortyFivesGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockFortyFivesGame) GetPhase() domain.FortyFivesPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FortyFivesPhase)
}

// IsHumanTurn モック
func (_m *MockFortyFivesGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockFortyFivesGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockFortyFivesGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockFortyFivesGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockFortyFivesGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockFortyFivesGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockFortyFivesGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockFortyFivesGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockFortyFivesGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockFortyFivesGame) GetContract() domain.FortyFivesBid {
	ret := _m.Called()
	return ret.Get(0).(domain.FortyFivesBid)
}

// GetBids モック
func (_m *MockFortyFivesGame) GetBids() [domain.FortyFivesPlayerCnt]domain.FortyFivesBid {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyFivesPlayerCnt]domain.FortyFivesBid)
}

// GetBidDone モック
func (_m *MockFortyFivesGame) GetBidDone() [domain.FortyFivesPlayerCnt]bool {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyFivesPlayerCnt]bool)
}

// GetTrumpSuit モック
func (_m *MockFortyFivesGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTeamScores モック
func (_m *MockFortyFivesGame) GetTeamScores() [domain.FortyFivesTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyFivesTeamCnt]int)
}

// GetRoundTeamPoints モック
func (_m *MockFortyFivesGame) GetRoundTeamPoints() [domain.FortyFivesTeamCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyFivesTeamCnt]int)
}

func (_m *MockFortyFivesGame) GetContractProgress() *domain.FortyFivesContractProgress {
	args := _m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.FortyFivesContractProgress)
}

// GetWinnerTeam モック
func (_m *MockFortyFivesGame) GetWinnerTeam() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockFortyFivesGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockFortyFivesGame) GetPlayer(i int) *domain.FortyFivesPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.FortyFivesPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockFortyFivesGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetTopTrumpIndices モック
func (_m *MockFortyFivesGame) GetTopTrumpIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]int)
}

// GetHint モック
func (_m *MockFortyFivesGame) GetHint() *domain.FortyFivesHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.FortyFivesHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockFortyFivesGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
