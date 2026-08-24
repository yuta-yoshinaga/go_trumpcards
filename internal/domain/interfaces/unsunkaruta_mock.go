//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockUnsunKarutaGame はうんすんカルタのゲームモック。
type MockUnsunKarutaGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockUnsunKarutaGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockUnsunKarutaGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockUnsunKarutaGame) PlayerPlay(cardIndex int, declare bool) error {
	return _m.Called(cardIndex, declare).Error(0)
}

// CpuPlay モック
func (_m *MockUnsunKarutaGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockUnsunKarutaGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockUnsunKarutaGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockUnsunKarutaGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockUnsunKarutaGame) GetConfig() domain.UnsunKarutaConfig {
	return _m.Called().Get(0).(domain.UnsunKarutaConfig)
}

// SetConfig モック
func (_m *MockUnsunKarutaGame) SetConfig(cfg domain.UnsunKarutaConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockUnsunKarutaGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockUnsunKarutaGame) GetPhase() domain.UnsunKarutaPhase {
	return _m.Called().Get(0).(domain.UnsunKarutaPhase)
}

// IsHumanTurn モック
func (_m *MockUnsunKarutaGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// CanDeclare モック
func (_m *MockUnsunKarutaGame) CanDeclare() bool { return _m.Called().Bool(0) }

// IsMustFollow モック
func (_m *MockUnsunKarutaGame) IsMustFollow() bool { return _m.Called().Bool(0) }

// IsDeclaredThisTrick モック
func (_m *MockUnsunKarutaGame) IsDeclaredThisTrick() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockUnsunKarutaGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockUnsunKarutaGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockUnsunKarutaGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockUnsunKarutaGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockUnsunKarutaGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockUnsunKarutaGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetTrumpSuit モック
func (_m *MockUnsunKarutaGame) GetTrumpSuit() int { return _m.Called().Int(0) }

// TrumpCard モック
func (_m *MockUnsunKarutaGame) TrumpCard() *domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

// GetTeamTricks モック
func (_m *MockUnsunKarutaGame) GetTeamTricks() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetTeamScores モック
func (_m *MockUnsunKarutaGame) GetTeamScores() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetLastTrickWinner モック
func (_m *MockUnsunKarutaGame) GetLastTrickWinner() int { return _m.Called().Int(0) }

// GetResult モック
func (_m *MockUnsunKarutaGame) GetResult() domain.UnsunKarutaResult {
	return _m.Called().Get(0).(domain.UnsunKarutaResult)
}

// GetWinnerTeam モック
func (_m *MockUnsunKarutaGame) GetWinnerTeam() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockUnsunKarutaGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockUnsunKarutaGame) GetPlayer(i int) *domain.UnsunKarutaPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.UnsunKarutaPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockUnsunKarutaGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockUnsunKarutaGame) GetHint() *domain.UnsunKarutaHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.UnsunKarutaHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockUnsunKarutaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
