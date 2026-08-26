//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPiedmonteseTarotGame はピエモンテ・タロッコのゲームモック。
type MockPiedmonteseTarotGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPiedmonteseTarotGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockPiedmonteseTarotGame) NextRound() { _m.Called() }

// PlayerScarto モック
func (_m *MockPiedmonteseTarotGame) PlayerScarto(cardIndices []int) error {
	return _m.Called(cardIndices).Error(0)
}

// CpuScarto モック
func (_m *MockPiedmonteseTarotGame) CpuScarto() { _m.Called() }

// PlayerPlay モック
func (_m *MockPiedmonteseTarotGame) PlayerPlay(cardIndex int) error {
	return _m.Called(cardIndex).Error(0)
}

// CpuPlay モック
func (_m *MockPiedmonteseTarotGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockPiedmonteseTarotGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockPiedmonteseTarotGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockPiedmonteseTarotGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockPiedmonteseTarotGame) GetConfig() domain.PiedmonteseTarotConfig {
	return _m.Called().Get(0).(domain.PiedmonteseTarotConfig)
}

// SetConfig モック
func (_m *MockPiedmonteseTarotGame) SetConfig(cfg domain.PiedmonteseTarotConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockPiedmonteseTarotGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockPiedmonteseTarotGame) GetPhase() domain.PiedmonteseTarotPhase {
	return _m.Called().Get(0).(domain.PiedmonteseTarotPhase)
}

// IsHumanTurn モック
func (_m *MockPiedmonteseTarotGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// IsHumanScartoTurn モック
func (_m *MockPiedmonteseTarotGame) IsHumanScartoTurn() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockPiedmonteseTarotGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockPiedmonteseTarotGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockPiedmonteseTarotGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockPiedmonteseTarotGame) GetCurrentTrick() []*domain.TrickCard {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.TrickCard); ok {
		return v
	}
	return nil
}

// GetLeadPlayerIdx モック
func (_m *MockPiedmonteseTarotGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockPiedmonteseTarotGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetScartoCount モック
func (_m *MockPiedmonteseTarotGame) GetScartoCount() int { return _m.Called().Int(0) }

// TalonSize モック
func (_m *MockPiedmonteseTarotGame) TalonSize() int { return _m.Called().Int(0) }

// HandSize モック
func (_m *MockPiedmonteseTarotGame) HandSize() int { return _m.Called().Int(0) }

// GetPlayerScores モック
func (_m *MockPiedmonteseTarotGame) GetPlayerScores() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetDealScores モック
func (_m *MockPiedmonteseTarotGame) GetDealScores() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetCardThirds モック
func (_m *MockPiedmonteseTarotGame) GetCardThirds(i int) int { return _m.Called(i).Int(0) }

// GetLastTrickWinner モック
func (_m *MockPiedmonteseTarotGame) GetLastTrickWinner() int { return _m.Called().Int(0) }

// GetOutcome モック
func (_m *MockPiedmonteseTarotGame) GetOutcome() domain.PiedmonteseTarotOutcome {
	return _m.Called().Get(0).(domain.PiedmonteseTarotOutcome)
}

// GetResult モック
func (_m *MockPiedmonteseTarotGame) GetResult() domain.PiedmonteseTarotResult {
	return _m.Called().Get(0).(domain.PiedmonteseTarotResult)
}

// GetWinnerPlayer モック
func (_m *MockPiedmonteseTarotGame) GetWinnerPlayer() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockPiedmonteseTarotGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockPiedmonteseTarotGame) GetPlayer(i int) *domain.PiedmonteseTarotPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.PiedmonteseTarotPlayer); ok {
		return v
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockPiedmonteseTarotGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetDiscardableIndices モック
func (_m *MockPiedmonteseTarotGame) GetDiscardableIndices() []int {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]int); ok {
		return v
	}
	return nil
}

// GetHint モック
func (_m *MockPiedmonteseTarotGame) GetHint() *domain.PiedmonteseTarotHint {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.PiedmonteseTarotHint); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockPiedmonteseTarotGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
