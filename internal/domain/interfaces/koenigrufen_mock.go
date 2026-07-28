//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKoenigrufenGame ケーニッヒルーフェン (Königrufen) のゲームモック
type MockKoenigrufenGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKoenigrufenGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockKoenigrufenGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockKoenigrufenGame) PlayerBid(bid domain.KoenigrufenBid) error {
	return _m.Called(bid).Error(0)
}

// PlayerPass モック
func (_m *MockKoenigrufenGame) PlayerPass() error { return _m.Called().Error(0) }

// CpuBid モック
func (_m *MockKoenigrufenGame) CpuBid() { _m.Called() }

// PlayerCallKing モック
func (_m *MockKoenigrufenGame) PlayerCallKing(suit int) error {
	return _m.Called(suit).Error(0)
}

// CpuCallKing モック
func (_m *MockKoenigrufenGame) CpuCallKing() { _m.Called() }

// PlayerDiscard モック
func (_m *MockKoenigrufenGame) PlayerDiscard(cardIndices []int) error {
	return _m.Called(cardIndices).Error(0)
}

// CpuDiscard モック
func (_m *MockKoenigrufenGame) CpuDiscard() { _m.Called() }

// PlayerPlay モック
func (_m *MockKoenigrufenGame) PlayerPlay(cardIndex int) error {
	return _m.Called(cardIndex).Error(0)
}

// CpuPlay モック
func (_m *MockKoenigrufenGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockKoenigrufenGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockKoenigrufenGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockKoenigrufenGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockKoenigrufenGame) GetConfig() domain.KoenigrufenConfig {
	return _m.Called().Get(0).(domain.KoenigrufenConfig)
}

// SetConfig モック
func (_m *MockKoenigrufenGame) SetConfig(cfg domain.KoenigrufenConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockKoenigrufenGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockKoenigrufenGame) GetPhase() domain.KoenigrufenPhase {
	return _m.Called().Get(0).(domain.KoenigrufenPhase)
}

// IsHumanTurn モック
func (_m *MockKoenigrufenGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// IsHumanBidTurn モック
func (_m *MockKoenigrufenGame) IsHumanBidTurn() bool { return _m.Called().Bool(0) }

// IsHumanCallTurn モック
func (_m *MockKoenigrufenGame) IsHumanCallTurn() bool { return _m.Called().Bool(0) }

// IsHumanDiscardTurn モック
func (_m *MockKoenigrufenGame) IsHumanDiscardTurn() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockKoenigrufenGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockKoenigrufenGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockKoenigrufenGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockKoenigrufenGame) GetCurrentTrick() []*domain.TrickCard {
	return _m.Called().Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockKoenigrufenGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockKoenigrufenGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetBidPlayerIdx モック
func (_m *MockKoenigrufenGame) GetBidPlayerIdx() int { return _m.Called().Int(0) }

// GetHighestBid モック
func (_m *MockKoenigrufenGame) GetHighestBid() domain.KoenigrufenBid {
	return _m.Called().Get(0).(domain.KoenigrufenBid)
}

// GetHighestBidder モック
func (_m *MockKoenigrufenGame) GetHighestBidder() int { return _m.Called().Int(0) }

// GetDeclarerIdx モック
func (_m *MockKoenigrufenGame) GetDeclarerIdx() int { return _m.Called().Int(0) }

// GetContract モック
func (_m *MockKoenigrufenGame) GetContract() domain.KoenigrufenBid {
	return _m.Called().Get(0).(domain.KoenigrufenBid)
}

// GetCalledKing モック
func (_m *MockKoenigrufenGame) GetCalledKing() int { return _m.Called().Int(0) }

// GetPartnerIdx モック
func (_m *MockKoenigrufenGame) GetPartnerIdx() int { return _m.Called().Int(0) }

// GetPartnerRevealed モック
func (_m *MockKoenigrufenGame) GetPartnerRevealed() bool { return _m.Called().Bool(0) }

// GetTalonCount モック
func (_m *MockKoenigrufenGame) GetTalonCount() int { return _m.Called().Int(0) }

// GetTalon モック
func (_m *MockKoenigrufenGame) GetTalon() []*domain.Card {
	return _m.Called().Get(0).([]*domain.Card)
}

// GetStashOwner モック
func (_m *MockKoenigrufenGame) GetStashOwner() int { return _m.Called().Int(0) }

// GetPlayerScores モック
func (_m *MockKoenigrufenGame) GetPlayerScores() [domain.KoenigrufenPlayerCnt]int {
	return _m.Called().Get(0).([domain.KoenigrufenPlayerCnt]int)
}

// GetCardPoints モック
func (_m *MockKoenigrufenGame) GetCardPoints(i int) int { return _m.Called(i).Int(0) }

// GetOutcome モック
func (_m *MockKoenigrufenGame) GetOutcome() domain.KoenigrufenOutcome {
	return _m.Called().Get(0).(domain.KoenigrufenOutcome)
}

// GetResult モック
func (_m *MockKoenigrufenGame) GetResult() domain.KoenigrufenResult {
	return _m.Called().Get(0).(domain.KoenigrufenResult)
}

// GetWinnerPlayer モック
func (_m *MockKoenigrufenGame) GetWinnerPlayer() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockKoenigrufenGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockKoenigrufenGame) GetPlayer(i int) *domain.KoenigrufenPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KoenigrufenPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockKoenigrufenGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockKoenigrufenGame) GetHint() *domain.KoenigrufenHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KoenigrufenHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockKoenigrufenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
