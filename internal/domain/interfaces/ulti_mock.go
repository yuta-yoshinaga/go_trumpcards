//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockUltiGame ウルティ (Ulti) のゲームモック
type MockUltiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockUltiGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockUltiGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockUltiGame) PlayerBid(contract domain.UltiContract, trumpSuit int) error {
	ret := _m.Called(contract, trumpSuit)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockUltiGame) CpuBid() {
	_m.Called()
}

// PlayerDiscard モック
func (_m *MockUltiGame) PlayerDiscard(cardIndices []int) error {
	ret := _m.Called(cardIndices)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockUltiGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockUltiGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockUltiGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockUltiGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockUltiGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockUltiGame) GetConfig() domain.UltiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiConfig)
}

// SetConfig モック
func (_m *MockUltiGame) SetConfig(cfg domain.UltiConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockUltiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockUltiGame) GetPhase() domain.UltiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiPhase)
}

// IsHumanTurn モック
func (_m *MockUltiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockUltiGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockUltiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockUltiGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockUltiGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockUltiGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockUltiGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockUltiGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockUltiGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockUltiGame) GetContract() domain.UltiContract {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiContract)
}

// GetTrumpSuit モック
func (_m *MockUltiGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTalonCount モック
func (_m *MockUltiGame) GetTalonCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTalonTaken モック
func (_m *MockUltiGame) GetTalonTaken() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetDiscardCount モック
func (_m *MockUltiGame) GetDiscardCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCoins モック
func (_m *MockUltiGame) GetPlayerCoins() [domain.UltiPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.UltiPlayerCnt]int)
}

// GetCardPoints モック
func (_m *MockUltiGame) GetCardPoints(i int) int {
	ret := _m.Called(i)
	return ret.Get(0).(int)
}

// GetOutcome モック
func (_m *MockUltiGame) GetOutcome() domain.UltiOutcome {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiOutcome)
}

// GetResult モック
func (_m *MockUltiGame) GetResult() domain.UltiResult {
	ret := _m.Called()
	return ret.Get(0).(domain.UltiResult)
}

// GetWinnerPlayer モック
func (_m *MockUltiGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockUltiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockUltiGame) GetPlayer(i int) *domain.UltiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.UltiPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockUltiGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockUltiGame) GetHint() *domain.UltiHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.UltiHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockUltiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
