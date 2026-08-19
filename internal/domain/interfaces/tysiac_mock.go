//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTysiacGame サウザンド (Tysiąc) のゲームモック
type MockTysiacGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockTysiacGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockTysiacGame) NextRound() {
	_m.Called()
}

// PlayerBid モック
func (_m *MockTysiacGame) PlayerBid(raise bool) error {
	ret := _m.Called(raise)
	return ret.Error(0)
}

// CpuBid モック
func (_m *MockTysiacGame) CpuBid() {
	_m.Called()
}

// PlayerDiscard モック
func (_m *MockTysiacGame) PlayerDiscard(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockTysiacGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockTysiacGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockTysiacGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockTysiacGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockTysiacGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockTysiacGame) GetConfig() domain.TysiacConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.TysiacConfig)
}

// SetConfig モック
func (_m *MockTysiacGame) SetConfig(cfg domain.TysiacConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockTysiacGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockTysiacGame) GetPhase() domain.TysiacPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TysiacPhase)
}

// IsHumanTurn モック
func (_m *MockTysiacGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// IsHumanBidTurn モック
func (_m *MockTysiacGame) IsHumanBidTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockTysiacGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockTysiacGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockTysiacGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockTysiacGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockTysiacGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockTysiacGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetForehandIdx モック
func (_m *MockTysiacGame) GetForehandIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeclarerIdx モック
func (_m *MockTysiacGame) GetDeclarerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetContract モック
func (_m *MockTysiacGame) GetContract() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentBid モック
func (_m *MockTysiacGame) GetCurrentBid() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockTysiacGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerScores モック
func (_m *MockTysiacGame) GetPlayerScores() [domain.TysiacPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TysiacPlayerCnt]int)
}

// GetRoundCardPoints モック
func (_m *MockTysiacGame) GetRoundCardPoints() [domain.TysiacPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TysiacPlayerCnt]int)
}

// GetRoundMarriage モック
func (_m *MockTysiacGame) GetRoundMarriage() [domain.TysiacPlayerCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.TysiacPlayerCnt]int)
}

// GetMarriageOptions モック
func (_m *MockTysiacGame) GetMarriageOptions(playerIdx int) []domain.TysiacMarriageOption {
	ret := _m.Called(playerIdx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]domain.TysiacMarriageOption)
}

// GetWinnerPlayer モック
func (_m *MockTysiacGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockTysiacGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockTysiacGame) GetPlayer(i int) *domain.TysiacPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.TysiacPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockTysiacGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockTysiacGame) GetHint() *domain.TysiacHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.TysiacHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockTysiacGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
