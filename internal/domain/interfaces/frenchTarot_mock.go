//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFrenchTarotGame フレンチタロット (French Tarot) のゲームモック
type MockFrenchTarotGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockFrenchTarotGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockFrenchTarotGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockFrenchTarotGame) PlayerBid(bid domain.FrenchTarotBid) error {
	return _m.Called(bid).Error(0)
}

// PlayerPass モック
func (_m *MockFrenchTarotGame) PlayerPass() error { return _m.Called().Error(0) }

// CpuBid モック
func (_m *MockFrenchTarotGame) CpuBid() { _m.Called() }

// PlayerDiscard モック
func (_m *MockFrenchTarotGame) PlayerDiscard(cardIndices []int) error {
	return _m.Called(cardIndices).Error(0)
}

// CpuDiscard モック
func (_m *MockFrenchTarotGame) CpuDiscard() { _m.Called() }

// PlayerPlay モック
func (_m *MockFrenchTarotGame) PlayerPlay(cardIndex int) error {
	return _m.Called(cardIndex).Error(0)
}

// CpuPlay モック
func (_m *MockFrenchTarotGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockFrenchTarotGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockFrenchTarotGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockFrenchTarotGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockFrenchTarotGame) GetConfig() domain.FrenchTarotConfig {
	return _m.Called().Get(0).(domain.FrenchTarotConfig)
}

// SetConfig モック
func (_m *MockFrenchTarotGame) SetConfig(cfg domain.FrenchTarotConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockFrenchTarotGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockFrenchTarotGame) GetPhase() domain.FrenchTarotPhase {
	return _m.Called().Get(0).(domain.FrenchTarotPhase)
}

// IsHumanTurn モック
func (_m *MockFrenchTarotGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// IsHumanBidTurn モック
func (_m *MockFrenchTarotGame) IsHumanBidTurn() bool { return _m.Called().Bool(0) }

// IsHumanDiscardTurn モック
func (_m *MockFrenchTarotGame) IsHumanDiscardTurn() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockFrenchTarotGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockFrenchTarotGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockFrenchTarotGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockFrenchTarotGame) GetCurrentTrick() []*domain.TrickCard {
	return _m.Called().Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockFrenchTarotGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockFrenchTarotGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetBidPlayerIdx モック
func (_m *MockFrenchTarotGame) GetBidPlayerIdx() int { return _m.Called().Int(0) }

// GetHighestBid モック
func (_m *MockFrenchTarotGame) GetHighestBid() domain.FrenchTarotBid {
	return _m.Called().Get(0).(domain.FrenchTarotBid)
}

// GetHighestBidder モック
func (_m *MockFrenchTarotGame) GetHighestBidder() int { return _m.Called().Int(0) }

// GetDeclarerIdx モック
func (_m *MockFrenchTarotGame) GetDeclarerIdx() int { return _m.Called().Int(0) }

// GetContract モック
func (_m *MockFrenchTarotGame) GetContract() domain.FrenchTarotBid {
	return _m.Called().Get(0).(domain.FrenchTarotBid)
}

// GetChienCount モック
func (_m *MockFrenchTarotGame) GetChienCount() int { return _m.Called().Int(0) }

// GetChien モック
func (_m *MockFrenchTarotGame) GetChien() []*domain.Card {
	return _m.Called().Get(0).([]*domain.Card)
}

// GetChienRevealed モック
func (_m *MockFrenchTarotGame) GetChienRevealed() bool { return _m.Called().Bool(0) }

// GetStashOwner モック
func (_m *MockFrenchTarotGame) GetStashOwner() int { return _m.Called().Int(0) }

// GetPlayerScores モック
func (_m *MockFrenchTarotGame) GetPlayerScores() [domain.FrenchTarotPlayerCnt]int {
	return _m.Called().Get(0).([domain.FrenchTarotPlayerCnt]int)
}

// GetCardPoints モック
func (_m *MockFrenchTarotGame) GetCardPoints(i int) int { return _m.Called(i).Int(0) }

// GetOutcome モック
func (_m *MockFrenchTarotGame) GetOutcome() domain.FrenchTarotOutcome {
	return _m.Called().Get(0).(domain.FrenchTarotOutcome)
}

// GetResult モック
func (_m *MockFrenchTarotGame) GetResult() domain.FrenchTarotResult {
	return _m.Called().Get(0).(domain.FrenchTarotResult)
}

// GetWinnerPlayer モック
func (_m *MockFrenchTarotGame) GetWinnerPlayer() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockFrenchTarotGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockFrenchTarotGame) GetPlayer(i int) *domain.FrenchTarotPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.FrenchTarotPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockFrenchTarotGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockFrenchTarotGame) GetHint() *domain.FrenchTarotHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.FrenchTarotHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockFrenchTarotGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
