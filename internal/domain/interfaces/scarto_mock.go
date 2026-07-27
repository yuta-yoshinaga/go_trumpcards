//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScartoGame スカルト (Scarto) のゲームモック
type MockScartoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockScartoGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockScartoGame) NextRound() { _m.Called() }

// PlayerScarto モック
func (_m *MockScartoGame) PlayerScarto(cardIndices []int) error {
	return _m.Called(cardIndices).Error(0)
}

// CpuScarto モック
func (_m *MockScartoGame) CpuScarto() { _m.Called() }

// PlayerPlay モック
func (_m *MockScartoGame) PlayerPlay(cardIndex int) error {
	return _m.Called(cardIndex).Error(0)
}

// CpuPlay モック
func (_m *MockScartoGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockScartoGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockScartoGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockScartoGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockScartoGame) GetConfig() domain.ScartoConfig {
	return _m.Called().Get(0).(domain.ScartoConfig)
}

// SetConfig モック
func (_m *MockScartoGame) SetConfig(cfg domain.ScartoConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockScartoGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockScartoGame) GetPhase() domain.ScartoPhase {
	return _m.Called().Get(0).(domain.ScartoPhase)
}

// IsHumanTurn モック
func (_m *MockScartoGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// IsHumanScartoTurn モック
func (_m *MockScartoGame) IsHumanScartoTurn() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockScartoGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockScartoGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockScartoGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockScartoGame) GetCurrentTrick() []*domain.TrickCard {
	return _m.Called().Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockScartoGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockScartoGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetScartoCount モック
func (_m *MockScartoGame) GetScartoCount() int { return _m.Called().Int(0) }

// GetPlayerScores モック
func (_m *MockScartoGame) GetPlayerScores() [domain.ScartoPlayerCnt]int {
	return _m.Called().Get(0).([domain.ScartoPlayerCnt]int)
}

// GetDealScores モック
func (_m *MockScartoGame) GetDealScores() [domain.ScartoPlayerCnt]int {
	return _m.Called().Get(0).([domain.ScartoPlayerCnt]int)
}

// GetCardPoints モック
func (_m *MockScartoGame) GetCardPoints(i int) int { return _m.Called(i).Int(0) }

// GetOutcome モック
func (_m *MockScartoGame) GetOutcome() domain.ScartoOutcome {
	return _m.Called().Get(0).(domain.ScartoOutcome)
}

// GetResult モック
func (_m *MockScartoGame) GetResult() domain.ScartoResult {
	return _m.Called().Get(0).(domain.ScartoResult)
}

// GetWinnerPlayer モック
func (_m *MockScartoGame) GetWinnerPlayer() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockScartoGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockScartoGame) GetPlayer(i int) *domain.ScartoPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.ScartoPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockScartoGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockScartoGame) GetHint() *domain.ScartoHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.ScartoHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockScartoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
