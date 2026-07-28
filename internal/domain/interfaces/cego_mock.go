//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCegoGame チェゴ (Cego) のゲームモック
type MockCegoGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCegoGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockCegoGame) NextRound() { _m.Called() }

// PlayerBid モック
func (_m *MockCegoGame) PlayerBid(bid domain.CegoBid) error {
	return _m.Called(bid).Error(0)
}

// PlayerPass モック
func (_m *MockCegoGame) PlayerPass() error { return _m.Called().Error(0) }

// CpuBid モック
func (_m *MockCegoGame) CpuBid() { _m.Called() }

// PlayerChooseContract モック
func (_m *MockCegoGame) PlayerChooseContract(ct domain.CegoContract) error {
	return _m.Called(ct).Error(0)
}

// CpuChooseContract モック
func (_m *MockCegoGame) CpuChooseContract() { _m.Called() }

// PlayerDiscard モック
func (_m *MockCegoGame) PlayerDiscard(keepIndices []int) error {
	return _m.Called(keepIndices).Error(0)
}

// CpuDiscard モック
func (_m *MockCegoGame) CpuDiscard() { _m.Called() }

// PlayerPlay モック
func (_m *MockCegoGame) PlayerPlay(cardIndex int) error {
	return _m.Called(cardIndex).Error(0)
}

// CpuPlay モック
func (_m *MockCegoGame) CpuPlay() { _m.Called() }

// ResolveTrick モック
func (_m *MockCegoGame) ResolveTrick() { _m.Called() }

// NextTrick モック
func (_m *MockCegoGame) NextTrick() { _m.Called() }

// ScoreRound モック
func (_m *MockCegoGame) ScoreRound() { _m.Called() }

// GetConfig モック
func (_m *MockCegoGame) GetConfig() domain.CegoConfig {
	return _m.Called().Get(0).(domain.CegoConfig)
}

// SetConfig モック
func (_m *MockCegoGame) SetConfig(cfg domain.CegoConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockCegoGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }

// GetPhase モック
func (_m *MockCegoGame) GetPhase() domain.CegoPhase {
	return _m.Called().Get(0).(domain.CegoPhase)
}

// IsHumanTurn モック
func (_m *MockCegoGame) IsHumanTurn() bool { return _m.Called().Bool(0) }

// IsHumanBidTurn モック
func (_m *MockCegoGame) IsHumanBidTurn() bool { return _m.Called().Bool(0) }

// IsHumanContractTurn モック
func (_m *MockCegoGame) IsHumanContractTurn() bool { return _m.Called().Bool(0) }

// IsHumanExchangeTurn モック
func (_m *MockCegoGame) IsHumanExchangeTurn() bool { return _m.Called().Bool(0) }

// GetRoundNumber モック
func (_m *MockCegoGame) GetRoundNumber() int { return _m.Called().Int(0) }

// GetTrickNumber モック
func (_m *MockCegoGame) GetTrickNumber() int { return _m.Called().Int(0) }

// GetCurrentPlayerIdx モック
func (_m *MockCegoGame) GetCurrentPlayerIdx() int { return _m.Called().Int(0) }

// GetCurrentTrick モック
func (_m *MockCegoGame) GetCurrentTrick() []*domain.TrickCard {
	return _m.Called().Get(0).([]*domain.TrickCard)
}

// GetLeadPlayerIdx モック
func (_m *MockCegoGame) GetLeadPlayerIdx() int { return _m.Called().Int(0) }

// GetDealerIdx モック
func (_m *MockCegoGame) GetDealerIdx() int { return _m.Called().Int(0) }

// GetBidPlayerIdx モック
func (_m *MockCegoGame) GetBidPlayerIdx() int { return _m.Called().Int(0) }

// GetHighestBid モック
func (_m *MockCegoGame) GetHighestBid() domain.CegoBid {
	return _m.Called().Get(0).(domain.CegoBid)
}

// GetHighestBidder モック
func (_m *MockCegoGame) GetHighestBidder() int { return _m.Called().Int(0) }

// GetDeclarerIdx モック
func (_m *MockCegoGame) GetDeclarerIdx() int { return _m.Called().Int(0) }

// GetContract モック
func (_m *MockCegoGame) GetContract() domain.CegoBid {
	return _m.Called().Get(0).(domain.CegoBid)
}

// GetContractType モック
func (_m *MockCegoGame) GetContractType() domain.CegoContract {
	return _m.Called().Get(0).(domain.CegoContract)
}

// GetBlindCount モック
func (_m *MockCegoGame) GetBlindCount() int { return _m.Called().Int(0) }

// GetStashOwner モック
func (_m *MockCegoGame) GetStashOwner() int { return _m.Called().Int(0) }

// GetPlayerScores モック
func (_m *MockCegoGame) GetPlayerScores() [domain.CegoPlayerCnt]int {
	return _m.Called().Get(0).([domain.CegoPlayerCnt]int)
}

// GetCardPoints モック
func (_m *MockCegoGame) GetCardPoints(i int) int { return _m.Called(i).Int(0) }

// GetOutcome モック
func (_m *MockCegoGame) GetOutcome() domain.CegoOutcome {
	return _m.Called().Get(0).(domain.CegoOutcome)
}

// GetResult モック
func (_m *MockCegoGame) GetResult() domain.CegoResult {
	return _m.Called().Get(0).(domain.CegoResult)
}

// GetWinnerPlayer モック
func (_m *MockCegoGame) GetWinnerPlayer() int { return _m.Called().Int(0) }

// GetPlayerCnt モック
func (_m *MockCegoGame) GetPlayerCnt() int { return _m.Called().Int(0) }

// GetPlayer モック
func (_m *MockCegoGame) GetPlayer(i int) *domain.CegoPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.CegoPlayer)
	}
	return nil
}

// GetPlayableIndices モック
func (_m *MockCegoGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetHint モック
func (_m *MockCegoGame) GetHint() *domain.CegoHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.CegoHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockCegoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
