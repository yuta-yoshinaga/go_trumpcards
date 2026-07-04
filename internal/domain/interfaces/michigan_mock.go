//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMichiganGame はミシガン (Michigan) のゲームモック。
type MockMichiganGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockMichiganGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockMichiganGame) NextRound() { _m.Called() }

// PlaceHumanBet モック
func (_m *MockMichiganGame) PlaceHumanBet(bets []int) error {
	ret := _m.Called(bets)
	return ret.Error(0)
}

// PlayCard モック
func (_m *MockMichiganGame) PlayCard(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

// GetConfig モック
func (_m *MockMichiganGame) GetConfig() domain.MichiganConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.MichiganConfig)
}

// SetConfig モック
func (_m *MockMichiganGame) SetConfig(cfg domain.MichiganConfig) { _m.Called(cfg) }

// GetPhase モック
func (_m *MockMichiganGame) GetPhase() domain.MichiganPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MichiganPhase)
}

// GetGameEndFlag モック
func (_m *MockMichiganGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetRoundNumber モック
func (_m *MockMichiganGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockMichiganGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockMichiganGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLeadPlayerIdx モック
func (_m *MockMichiganGame) GetLeadPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetAnte モック
func (_m *MockMichiganGame) GetAnte() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBetBudget モック
func (_m *MockMichiganGame) GetBetBudget() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetHumanBetPlaced モック
func (_m *MockMichiganGame) GetHumanBetPlaced() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetBoodleCnt モック
func (_m *MockMichiganGame) GetBoodleCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetBoodle モック
func (_m *MockMichiganGame) GetBoodle(i int) *domain.MichiganBoodle {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MichiganBoodle)
	}
	return nil
}

// GetSeqSuit モック
func (_m *MockMichiganGame) GetSeqSuit() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetSeqHighValue モック
func (_m *MockMichiganGame) GetSeqHighValue() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDeadHandCount モック
func (_m *MockMichiganGame) GetDeadHandCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinnerIdx モック
func (_m *MockMichiganGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetMatchWinnerIdx モック
func (_m *MockMichiganGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetResult モック
func (_m *MockMichiganGame) GetResult() domain.MichiganResult {
	ret := _m.Called()
	return ret.Get(0).(domain.MichiganResult)
}

// GetPlayerCnt モック
func (_m *MockMichiganGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockMichiganGame) GetPlayer(i int) *domain.MichiganPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.MichiganPlayer)
	}
	return nil
}

// GetChips モック
func (_m *MockMichiganGame) GetChips() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsHumanTurn モック
func (_m *MockMichiganGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPlayableIndices モック
func (_m *MockMichiganGame) GetPlayableIndices() []int {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockMichiganGame) GetHint() *domain.MichiganHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.MichiganHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockMichiganGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
