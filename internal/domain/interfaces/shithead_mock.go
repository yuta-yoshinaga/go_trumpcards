//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockShitheadGame シットヘッドゲームモック
type MockShitheadGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockShitheadGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockShitheadGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockShitheadGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerPlay モック
func (_m *MockShitheadGame) PlayerPlay(indices []int) error {
	ret := _m.Called(indices)
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockShitheadGame) CpuPlay() {
	_m.Called()
}

// SetConfig モック
func (_m *MockShitheadGame) SetConfig(config domain.ShitheadConfig) {
	_m.Called(config)
}

// GetPlayerCnt モック
func (_m *MockShitheadGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockShitheadGame) GetPlayer(i int) *domain.ShitheadPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.ShitheadPlayer); ok {
		return val
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockShitheadGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDiscardPile モック
func (_m *MockShitheadGame) GetDiscardPile() []*domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.Card); ok {
		return val
	}
	return nil
}

// GetTopCard モック
func (_m *MockShitheadGame) GetTopCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

// GetStockSize モック
func (_m *MockShitheadGame) GetStockSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetCpuActions モック
func (_m *MockShitheadGame) GetCpuActions() []*domain.ShitheadCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ShitheadCpuAction); ok {
		return val
	}
	return nil
}

// GetHumanAction モック
func (_m *MockShitheadGame) GetHumanAction() *domain.ShitheadCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.ShitheadCpuAction); ok {
		return val
	}
	return nil
}

// GetSkipNext モック
func (_m *MockShitheadGame) GetSkipNext() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetSevenActive モック
func (_m *MockShitheadGame) GetSevenActive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetConfig モック
func (_m *MockShitheadGame) GetConfig() domain.ShitheadConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.ShitheadConfig); ok {
		return val
	}
	return domain.ShitheadConfig{}
}

// CurrentSource モック
func (_m *MockShitheadGame) CurrentSource() string {
	ret := _m.Called()
	return ret.String(0)
}

// GetActionLog モック
func (_m *MockShitheadGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}
