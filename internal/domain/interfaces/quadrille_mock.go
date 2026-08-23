//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockQuadrilleGame カドリール (Quadrille) のゲームモック
type MockQuadrilleGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockQuadrilleGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockQuadrilleGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockQuadrilleGame) PlayerBid(bid domain.QuadrilleBid, trumpSuit int) error {
	ret := _m.Called(bid, trumpSuit)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockQuadrilleGame) CpuBid() {
	_m.Called()
}

// DeclareKing モック
func (_m *MockQuadrilleGame) DeclareKing(playerIdx, suit int) error {
	ret := _m.Called(playerIdx, suit)
	return ret.Error(0)
}

// CpuDeclareKing モック
func (_m *MockQuadrilleGame) CpuDeclareKing() {
	_m.Called()
}

// IsHumanKingCallTurn モック
func (_m *MockQuadrilleGame) IsHumanKingCallTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCalledKingSuit モック
func (_m *MockQuadrilleGame) GetCalledKingSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPartnerIdx モック
func (_m *MockQuadrilleGame) GetPartnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsRoiSeul モック
func (_m *MockQuadrilleGame) IsRoiSeul() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCallableKingSuits モック
func (_m *MockQuadrilleGame) GetCallableKingSuits() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetSideTrickCounts モック
func (_m *MockQuadrilleGame) GetSideTrickCounts() (int, int) {
	ret := _m.Called()
	return ret.Get(0).(int), ret.Get(1).(int)
}

// PlayerPlay モック
func (_m *MockQuadrilleGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockQuadrilleGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockQuadrilleGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockQuadrilleGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockQuadrilleGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockQuadrilleGame) GetConfig() domain.QuadrilleConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleConfig)
}

// SetConfig モック
func (_m *MockQuadrilleGame) SetConfig(cfg domain.QuadrilleConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockQuadrilleGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockQuadrilleGame) GetPhase() domain.QuadrillePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrillePhase)
}

// IsHumanTurn モック
func (_m *MockQuadrilleGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockQuadrilleGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockQuadrilleGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockQuadrilleGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockQuadrilleGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockQuadrilleGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockQuadrilleGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockQuadrilleGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetForehandIdx モック
func (_m *MockQuadrilleGame) GetForehandIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetQuadrilleIdx モック
func (_m *MockQuadrilleGame) GetQuadrilleIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHighestBid モック
func (_m *MockQuadrilleGame) GetHighestBid() domain.QuadrilleBid {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleBid)
}

// GetWinningBid モック
func (_m *MockQuadrilleGame) GetWinningBid() domain.QuadrilleBid {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleBid)
}

// GetTrumpSuit モック
func (_m *MockQuadrilleGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBidderIdx モック
func (_m *MockQuadrilleGame) GetCurrentBidderIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockQuadrilleGame) GetPlayerScores() [domain.QuadrillePlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.QuadrillePlayerCnt]int)
}

// GetOutcome モック
func (_m *MockQuadrilleGame) GetOutcome() domain.QuadrilleOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleOutcome)
}

// GetResult モック
func (_m *MockQuadrilleGame) GetResult() domain.QuadrilleResult {
	ret := _m.Called()
	return ret.Get(0).(domain.QuadrilleResult)
}

// GetWinnerPlayer モック
func (_m *MockQuadrilleGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockQuadrilleGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockQuadrilleGame) GetPlayer(i int) *domain.QuadrillePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.QuadrillePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockQuadrilleGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockQuadrilleGame) GetHint() *domain.QuadrilleHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.QuadrilleHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockQuadrilleGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
