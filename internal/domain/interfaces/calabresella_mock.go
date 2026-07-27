//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCalabresellaGame カラブレセッラ (Calabresella) のゲームモック
type MockCalabresellaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCalabresellaGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockCalabresellaGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockCalabresellaGame) PlayerBid(bid domain.CalabresellaBid) error {
	ret := _m.Called(bid)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockCalabresellaGame) CpuBid() {
	_m.Called()
}

// PlayerDiscard モック
func (_m *MockCalabresellaGame) PlayerDiscard(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockCalabresellaGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockCalabresellaGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockCalabresellaGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockCalabresellaGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockCalabresellaGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockCalabresellaGame) GetConfig() domain.CalabresellaConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.CalabresellaConfig)
}

// SetConfig モック
func (_m *MockCalabresellaGame) SetConfig(cfg domain.CalabresellaConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockCalabresellaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockCalabresellaGame) GetPhase() domain.CalabresellaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CalabresellaPhase)
}

// IsHumanTurn モック
func (_m *MockCalabresellaGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockCalabresellaGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockCalabresellaGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockCalabresellaGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockCalabresellaGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockCalabresellaGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockCalabresellaGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockCalabresellaGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetForehandIdx モック
func (_m *MockCalabresellaGame) GetForehandIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSoloistIdx モック
func (_m *MockCalabresellaGame) GetSoloistIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinningBid モック
func (_m *MockCalabresellaGame) GetWinningBid() domain.CalabresellaBid {
	ret := _m.Called()
	return ret.Get(0).(domain.CalabresellaBid)
}

// GetCurrentBidderIdx モック
func (_m *MockCalabresellaGame) GetCurrentBidderIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockCalabresellaGame) GetPlayerScores() [domain.CalabresellaPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.CalabresellaPlayerCnt]int)
}

// GetRoundThirds モック
func (_m *MockCalabresellaGame) GetRoundThirds() [domain.CalabresellaPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.CalabresellaPlayerCnt]int)
}

// GetWinnerPlayer モック
func (_m *MockCalabresellaGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockCalabresellaGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockCalabresellaGame) GetPlayer(i int) *domain.CalabresellaPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.CalabresellaPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockCalabresellaGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockCalabresellaGame) GetHint() *domain.CalabresellaHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.CalabresellaHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockCalabresellaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
