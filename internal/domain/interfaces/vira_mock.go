//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockViraGame ヴィーラのゲームモック
type MockViraGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockViraGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockViraGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockViraGame) PlayerBid(bid domain.ViraBid) error {
	args := _m.Called(bid)
	return args.Error(0)
}

// CpuBid モック
func (_m *MockViraGame) CpuBid() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockViraGame) PlayerPlay(cardIndex int) error {
	args := _m.Called(cardIndex)
	return args.Error(0)
}

// CpuPlay モック
func (_m *MockViraGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockViraGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockViraGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockViraGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockViraGame) GetConfig() domain.ViraConfig {
	args := _m.Called()
	return args.Get(0).(domain.ViraConfig)
}

// SetConfig モック
func (_m *MockViraGame) SetConfig(cfg domain.ViraConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockViraGame) GetGameEndFlag() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetPhase モック
func (_m *MockViraGame) GetPhase() domain.ViraPhase {
	args := _m.Called()
	return args.Get(0).(domain.ViraPhase)
}

// IsHumanTurn モック
func (_m *MockViraGame) IsHumanTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// IsHumanBidTurn モック
func (_m *MockViraGame) IsHumanBidTurn() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetRoundNumber モック
func (_m *MockViraGame) GetRoundNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTrickNumber モック
func (_m *MockViraGame) GetTrickNumber() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentPlayerIdx モック
func (_m *MockViraGame) GetCurrentPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetCurrentTrick モック
func (_m *MockViraGame) GetCurrentTrick() []*domain.TrickCard {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockViraGame) GetLeadPlayerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDealerIdx モック
func (_m *MockViraGame) GetDealerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetDeclarerIdx モック
func (_m *MockViraGame) GetDeclarerIdx() int {
	args := _m.Called()
	return args.Int(0)
}

// GetContract モック
func (_m *MockViraGame) GetContract() domain.ViraBid {
	args := _m.Called()
	return args.Get(0).(domain.ViraBid)
}

// GetBids モック
func (_m *MockViraGame) GetBids() [domain.ViraPlayerCnt]domain.ViraBid {
	args := _m.Called()
	return args.Get(0).([domain.ViraPlayerCnt]domain.ViraBid)
}

// GetBidDone モック
func (_m *MockViraGame) GetBidDone() [domain.ViraPlayerCnt]bool {
	args := _m.Called()
	return args.Get(0).([domain.ViraPlayerCnt]bool)
}

// GetTrumpSuit モック
func (_m *MockViraGame) GetTrumpSuit() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPot モック
func (_m *MockViraGame) GetPot() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerScores モック
func (_m *MockViraGame) GetPlayerScores() [domain.ViraPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.ViraPlayerCnt]int)
}

// GetRoundTricks モック
func (_m *MockViraGame) GetRoundTricks() [domain.ViraPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.ViraPlayerCnt]int)
}

// GetLastRoundDelta モック
func (_m *MockViraGame) GetLastRoundDelta() [domain.ViraPlayerCnt]int {
	args := _m.Called()
	return args.Get(0).([domain.ViraPlayerCnt]int)
}

// GetLastRoundMade モック
func (_m *MockViraGame) GetLastRoundMade() bool {
	args := _m.Called()
	return args.Bool(0)
}

// GetWinnerPlayer モック
func (_m *MockViraGame) GetWinnerPlayer() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayerCnt モック
func (_m *MockViraGame) GetPlayerCnt() int {
	args := _m.Called()
	return args.Int(0)
}

// GetPlayer モック
func (_m *MockViraGame) GetPlayer(i int) *domain.ViraPlayer {
	args := _m.Called(i)
	if v, ok := args.Get(0).(*domain.ViraPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockViraGame) GetPlayableIndices(playerIdx int) []int {
	args := _m.Called(playerIdx)
	if v, ok := args.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockViraGame) GetHint() *domain.ViraHint {
	args := _m.Called()
	if v, ok := args.Get(0).(*domain.ViraHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockViraGame) GetActionLog() []*domain.ActionLogEntry {
	args := _m.Called()
	if v, ok := args.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
