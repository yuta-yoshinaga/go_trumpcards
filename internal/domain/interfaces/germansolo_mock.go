//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGermanSoloGame ジャーマン・ソロ (GermanSolo) のゲームモック
type MockGermanSoloGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGermanSoloGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockGermanSoloGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockGermanSoloGame) PlayerBid(bid domain.GermanSoloBid, trumpSuit int) error {
	ret := _m.Called(bid, trumpSuit)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockGermanSoloGame) CpuBid() {
	_m.Called()
}

// DeclareAce モック
func (_m *MockGermanSoloGame) DeclareAce(playerIdx, suit int) error {
	ret := _m.Called(playerIdx, suit)
	return ret.Error(0)
}

// CpuDeclareAce モック
func (_m *MockGermanSoloGame) CpuDeclareAce() {
	_m.Called()
}

// IsHumanAceCallTurn モック
func (_m *MockGermanSoloGame) IsHumanAceCallTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCalledAceSuit モック
func (_m *MockGermanSoloGame) GetCalledAceSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPartnerIdx モック
func (_m *MockGermanSoloGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsPlayingAlone モック
func (_m *MockGermanSoloGame) IsPlayingAlone() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCallableAceSuits モック
func (_m *MockGermanSoloGame) GetCallableAceSuits() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetSideTrickCounts モック
func (_m *MockGermanSoloGame) GetSideTrickCounts() (int, int) {
	ret := _m.Called()
	return ret.Get(0).(int), ret.Get(1).(int)
}

// PlayerPlay モック
func (_m *MockGermanSoloGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockGermanSoloGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockGermanSoloGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockGermanSoloGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockGermanSoloGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockGermanSoloGame) GetConfig() domain.GermanSoloConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloConfig)
}

// SetConfig モック
func (_m *MockGermanSoloGame) SetConfig(cfg domain.GermanSoloConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockGermanSoloGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockGermanSoloGame) GetPhase() domain.GermanSoloPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloPhase)
}

// IsHumanTurn モック
func (_m *MockGermanSoloGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockGermanSoloGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockGermanSoloGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockGermanSoloGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockGermanSoloGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockGermanSoloGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockGermanSoloGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockGermanSoloGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetForehandIdx モック
func (_m *MockGermanSoloGame) GetForehandIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockGermanSoloGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHighestBid モック
func (_m *MockGermanSoloGame) GetHighestBid() domain.GermanSoloBid {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloBid)
}

// GetBiddableBids モック
func (_m *MockGermanSoloGame) GetBiddableBids() []int {
	args := _m.Called()
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// RequiredTricks モック
func (_m *MockGermanSoloGame) RequiredTricks() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDeclarerSideSize モック
func (_m *MockGermanSoloGame) GetDeclarerSideSize() int {
	args := _m.Called()
	return args.Int(0)
}

// GetWinningBid モック
func (_m *MockGermanSoloGame) GetWinningBid() domain.GermanSoloBid {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloBid)
}

// GetTrumpSuit モック
func (_m *MockGermanSoloGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBidderIdx モック
func (_m *MockGermanSoloGame) GetCurrentBidderIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockGermanSoloGame) GetPlayerScores() [domain.GermanSoloPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.GermanSoloPlayerCnt]int)
}

// GetOutcome モック
func (_m *MockGermanSoloGame) GetOutcome() domain.GermanSoloOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloOutcome)
}

// GetResult モック
func (_m *MockGermanSoloGame) GetResult() domain.GermanSoloResult {
	ret := _m.Called()
	return ret.Get(0).(domain.GermanSoloResult)
}

// GetWinnerPlayer モック
func (_m *MockGermanSoloGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockGermanSoloGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockGermanSoloGame) GetPlayer(i int) *domain.GermanSoloPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.GermanSoloPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockGermanSoloGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockGermanSoloGame) GetHint() *domain.GermanSoloHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.GermanSoloHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockGermanSoloGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
