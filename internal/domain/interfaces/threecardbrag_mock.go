//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeCardBragGame スリーカード・ブラグのゲームモック
type MockThreeCardBragGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockThreeCardBragGame) Reset() {
	_m.Called()
}

// NextRound モック
func (_m *MockThreeCardBragGame) NextRound() {
	_m.Called()
}

// PlayerSee モック
func (_m *MockThreeCardBragGame) PlayerSee() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerBet モック
func (_m *MockThreeCardBragGame) PlayerBet() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerRaise モック
func (_m *MockThreeCardBragGame) PlayerRaise(newStake int) error {
	ret := _m.Called(newStake)
	return ret.Error(0)
}

// PlayerFold モック
func (_m *MockThreeCardBragGame) PlayerFold() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerShow モック
func (_m *MockThreeCardBragGame) PlayerShow() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuAct モック
func (_m *MockThreeCardBragGame) CpuAct() {
	_m.Called()
}

// GetConfig モック
func (_m *MockThreeCardBragGame) GetConfig() domain.ThreeCardBragConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.ThreeCardBragConfig)
}

// SetConfig モック
func (_m *MockThreeCardBragGame) SetConfig(cfg domain.ThreeCardBragConfig) {
	_m.Called(cfg)
}

// GetPhase モック
func (_m *MockThreeCardBragGame) GetPhase() domain.ThreeCardBragPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ThreeCardBragPhase)
}

// GetRoundNumber モック
func (_m *MockThreeCardBragGame) GetRoundNumber() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetCurrentPlayerIdx モック
func (_m *MockThreeCardBragGame) GetCurrentPlayerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetDealerIdx モック
func (_m *MockThreeCardBragGame) GetDealerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPot モック
func (_m *MockThreeCardBragGame) GetPot() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetStake モック
func (_m *MockThreeCardBragGame) GetStake() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetRoundWinnerIdx モック
func (_m *MockThreeCardBragGame) GetRoundWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// IsShowdown モック
func (_m *MockThreeCardBragGame) IsShowdown() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetGameEndFlag モック
func (_m *MockThreeCardBragGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetMatchWinnerIdx モック
func (_m *MockThreeCardBragGame) GetMatchWinnerIdx() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayerCnt モック
func (_m *MockThreeCardBragGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

// GetPlayer モック
func (_m *MockThreeCardBragGame) GetPlayer(i int) *domain.ThreeCardBragPlayer {
	ret := _m.Called(i)
	if v := ret.Get(0); v != nil {
		return v.(*domain.ThreeCardBragPlayer)
	}
	return nil
}

// IsHumanTurn モック
func (_m *MockThreeCardBragGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// CanShow モック
func (_m *MockThreeCardBragGame) CanShow() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetHint モック
func (_m *MockThreeCardBragGame) GetHint() *domain.ThreeCardBragHint {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.(*domain.ThreeCardBragHint)
	}
	return nil
}

// GetActionLog モック
func (_m *MockThreeCardBragGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v := ret.Get(0); v != nil {
		return v.([]*domain.ActionLogEntry)
	}
	return nil
}
