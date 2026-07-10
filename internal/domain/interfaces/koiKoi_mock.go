//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKoiKoiGame はこいこい (Koi-Koi) のゲームモック。
type MockKoiKoiGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockKoiKoiGame) Reset() { _m.Called() }

// NextRound モック
func (_m *MockKoiKoiGame) NextRound() { _m.Called() }

// PlayerPlay モック
func (_m *MockKoiKoiGame) PlayerPlay(handIdx, fieldIdx int) error {
	ret := _m.Called(handIdx, fieldIdx)
	return ret.Error(0)
}

// PlayerDecide モック
func (_m *MockKoiKoiGame) PlayerDecide(koikoi bool) error {
	ret := _m.Called(koikoi)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockKoiKoiGame) CpuPlay() { _m.Called() }

// CpuDecide モック
func (_m *MockKoiKoiGame) CpuDecide() { _m.Called() }

// GetConfig モック
func (_m *MockKoiKoiGame) GetConfig() domain.KoiKoiConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.KoiKoiConfig)
}

// SetConfig モック
func (_m *MockKoiKoiGame) SetConfig(cfg domain.KoiKoiConfig) { _m.Called(cfg) }

// GetGameEndFlag モック
func (_m *MockKoiKoiGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetPhase モック
func (_m *MockKoiKoiGame) GetPhase() domain.KoiKoiPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KoiKoiPhase)
}

// IsHumanTurn モック
func (_m *MockKoiKoiGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetCurrentTurn モック
func (_m *MockKoiKoiGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetFieldCards モック
func (_m *MockKoiKoiGame) GetFieldCards() []*domain.Card {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.Card)
	}
	return nil
}

// GetRemainingDeck モック
func (_m *MockKoiKoiGame) GetRemainingDeck() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundNumber モック
func (_m *MockKoiKoiGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetKoikoiCount モック
func (_m *MockKoiKoiGame) GetKoikoiCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinner モック
func (_m *MockKoiKoiGame) GetRoundWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetLastRoundResult モック
func (_m *MockKoiKoiGame) GetLastRoundResult() *domain.KoiKoiRoundResult {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KoiKoiRoundResult)
	}
	return nil
}

// GetPendingYaku モック
func (_m *MockKoiKoiGame) GetPendingYaku() []domain.KoiKoiYaku {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]domain.KoiKoiYaku)
	}
	return nil
}

// GetPendingPoints モック
func (_m *MockKoiKoiGame) GetPendingPoints() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetWinner モック
func (_m *MockKoiKoiGame) GetWinner() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockKoiKoiGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockKoiKoiGame) GetPlayer(i int) *domain.KoiKoiPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.KoiKoiPlayer)
	}
	return nil
}

// GetYaku モック
func (_m *MockKoiKoiGame) GetYaku(playerIdx int) ([]domain.KoiKoiYaku, int) {
	ret := _m.Called(playerIdx)
	var yakus []domain.KoiKoiYaku
	if v := ret.Get(0); v != nil {
		yakus = v.([]domain.KoiKoiYaku)
	}
	return yakus, ret.Int(1)
}

// GetPlayableIndices モック
func (_m *MockKoiKoiGame) GetPlayableIndices(playerIdx int) []int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.([]int)
	}
	return nil
}

// GetCaptureOptions モック
func (_m *MockKoiKoiGame) GetCaptureOptions(playerIdx int) map[int][]int {
	ret := _m.Called(playerIdx)
	if v := ret.Get(0); v != nil {
		return v.(map[int][]int)
	}
	return nil
}

// GetHint モック
func (_m *MockKoiKoiGame) GetHint() *domain.KoiKoiHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.KoiKoiHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockKoiKoiGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
