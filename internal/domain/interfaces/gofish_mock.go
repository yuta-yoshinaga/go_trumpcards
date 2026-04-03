//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGoFishGame Go Fishゲームモック
type MockGoFishGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockGoFishGame) Reset() {
	_m.Called()
}

// SetConfig モック
func (_m *MockGoFishGame) SetConfig(config domain.GoFishConfig) {
	_m.Called(config)
}

// GetGameEndFlag モック
func (_m *MockGoFishGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockGoFishGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerAsk モック
func (_m *MockGoFishGame) PlayerAsk(targetIdx, rank int) error {
	ret := _m.Called(targetIdx, rank)
	return ret.Error(0)
}

// CpuAsk モック
func (_m *MockGoFishGame) CpuAsk() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetPlayerCnt モック
func (_m *MockGoFishGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockGoFishGame) GetPlayer(i int) *domain.GoFishPlayer {
	ret := _m.Called(i)
	if v, ok := ret.Get(0).(*domain.GoFishPlayer); ok {
		return v
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockGoFishGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPhase モック
func (_m *MockGoFishGame) GetPhase() domain.GoFishPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GoFishPhase)
}

// GetConfig モック
func (_m *MockGoFishGame) GetConfig() domain.GoFishConfig {
	ret := _m.Called()
	return ret.Get(0).(domain.GoFishConfig)
}

// GetWinnerIdx モック
func (_m *MockGoFishGame) GetWinnerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTurnNumber モック
func (_m *MockGoFishGame) GetTurnNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDeckRemaining モック
func (_m *MockGoFishGame) GetDeckRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastAskPlayerIdx モック
func (_m *MockGoFishGame) GetLastAskPlayerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastAskTargetIdx モック
func (_m *MockGoFishGame) GetLastAskTargetIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastAskRank モック
func (_m *MockGoFishGame) GetLastAskRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLastAskSuccess モック
func (_m *MockGoFishGame) GetLastAskSuccess() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastCardsReceived モック
func (_m *MockGoFishGame) GetLastCardsReceived() []*domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.Card); ok {
		return v
	}
	return nil
}

// GetLastDrawnCard モック
func (_m *MockGoFishGame) GetLastDrawnCard() *domain.Card {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.Card); ok {
		return v
	}
	return nil
}

// GetLastBookFormed モック
func (_m *MockGoFishGame) GetLastBookFormed() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetLastBookRank モック
func (_m *MockGoFishGame) GetLastBookRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCpuActions モック
func (_m *MockGoFishGame) GetCpuActions() []*domain.GoFishCpuAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.GoFishCpuAction); ok {
		return v
	}
	return nil
}

// GetHumanAction モック
func (_m *MockGoFishGame) GetHumanAction() *domain.GoFishCpuAction {
	ret := _m.Called()
	if v, ok := ret.Get(0).(*domain.GoFishCpuAction); ok {
		return v
	}
	return nil
}

// GetActionLog モック
func (_m *MockGoFishGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if v, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return v
	}
	return nil
}
