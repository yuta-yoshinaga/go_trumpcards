//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpoilFiveGame スポイル・ファイブのゲームモック
type MockSpoilFiveGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockSpoilFiveGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockSpoilFiveGame) NextRound() {
	_m.Called()
}

// PlayerPlay モック
func (_m *MockSpoilFiveGame) PlayerPlay(cardIndex int) error {
	ret := _m.Called(cardIndex)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockSpoilFiveGame) CpuPlay() {
	_m.Called()
}

// ResolveTrick モック
func (_m *MockSpoilFiveGame) ResolveTrick() {
	_m.Called()
}

// NextTrick モック
func (_m *MockSpoilFiveGame) NextTrick() {
	_m.Called()
}

// ScoreRound モック
func (_m *MockSpoilFiveGame) ScoreRound() {
	_m.Called()
}

// GetConfig モック
func (_m *MockSpoilFiveGame) GetConfig() domain.SpoilFiveConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoilFiveConfig)
}

// SetConfig モック
func (_m *MockSpoilFiveGame) SetConfig(cfg domain.SpoilFiveConfig) {
	_m.Called(cfg)
}

// GetGameEndFlag モック
func (_m *MockSpoilFiveGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockSpoilFiveGame) GetPhase() domain.SpoilFivePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SpoilFivePhase)
}

// IsHumanTurn モック
func (_m *MockSpoilFiveGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockSpoilFiveGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrickNumber モック
func (_m *MockSpoilFiveGame) GetTrickNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockSpoilFiveGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentTrick モック
func (_m *MockSpoilFiveGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.TrickCard)
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockSpoilFiveGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockSpoilFiveGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetTrumpSuit モック
func (_m *MockSpoilFiveGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockSpoilFiveGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinnerIdx モック
func (_m *MockSpoilFiveGame) GetRoundWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerPlayer モック
func (_m *MockSpoilFiveGame) GetWinnerPlayer() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockSpoilFiveGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockSpoilFiveGame) GetPlayer(i int) *domain.SpoilFivePlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.SpoilFivePlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockSpoilFiveGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockSpoilFiveGame) GetHint() *domain.SpoilFiveHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.SpoilFiveHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockSpoilFiveGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
