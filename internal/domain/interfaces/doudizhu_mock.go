//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDoudizhuGame 斗地主ゲームモック
type MockDoudizhuGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDoudizhuGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockDoudizhuGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockDoudizhuGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerBid モック
func (_m *MockDoudizhuGame) PlayerBid(value int) error {
	ret := _m.Called(value)
	return ret.Error(0)
}

// PlayerPlay モック
func (_m *MockDoudizhuGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockDoudizhuGame) CpuPlay() {
	_m.Called()
}

// HasPendingAction モック
func (_m *MockDoudizhuGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockDoudizhuGame) SetConfig(config domain.DoudizhuConfig) {
	_m.Called(config)
}

// GetPhase モック
func (_m *MockDoudizhuGame) GetPhase() domain.DoudizhuPhase {
	ret := _m.Called()
	return domain.DoudizhuPhase(ret.Int(0))
}

// GetPlayerCnt モック
func (_m *MockDoudizhuGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockDoudizhuGame) GetPlayer(i int) *domain.DoudizhuPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.DoudizhuPlayer); ok {
		return val
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockDoudizhuGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTableCombo モック
func (_m *MockDoudizhuGame) GetTableCombo() *domain.DoudizhuCombo {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoudizhuCombo); ok {
		return val
	}
	return nil
}

// GetLastPlayIdx モック
func (_m *MockDoudizhuGame) GetLastPlayIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetKittyCards モック
func (_m *MockDoudizhuGame) GetKittyCards() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetLandlordIdx モック
func (_m *MockDoudizhuGame) GetLandlordIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBaseBid モック
func (_m *MockDoudizhuGame) GetBaseBid() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetBombCount モック
func (_m *MockDoudizhuGame) GetBombCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetScores モック
func (_m *MockDoudizhuGame) GetScores() [domain.DoudizhuPlayerCnt]int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([domain.DoudizhuPlayerCnt]int); ok {
		return val
	}
	return [domain.DoudizhuPlayerCnt]int{}
}

// GetBidValues モック
func (_m *MockDoudizhuGame) GetBidValues() [domain.DoudizhuPlayerCnt]int {
	ret := _m.Called()
	if val, ok := ret.Get(0).([domain.DoudizhuPlayerCnt]int); ok {
		return val
	}
	return [domain.DoudizhuPlayerCnt]int{}
}

// GetHighestBid モック
func (_m *MockDoudizhuGame) GetHighestBid() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCpuActions モック
func (_m *MockDoudizhuGame) GetCpuActions() []*domain.DoudizhuCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DoudizhuCpuAction); ok {
		return val
	}
	return nil
}

// GetHumanAction モック
func (_m *MockDoudizhuGame) GetHumanAction() *domain.DoudizhuCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DoudizhuCpuAction); ok {
		return val
	}
	return nil
}

// GetConfig モック
func (_m *MockDoudizhuGame) GetConfig() domain.DoudizhuConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.DoudizhuConfig); ok {
		return val
	}
	return domain.DoudizhuConfig{}
}

// GetActionLog モック
func (_m *MockDoudizhuGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}
