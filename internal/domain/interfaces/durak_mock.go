//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDurakGame ドゥラークゲームモック
type MockDurakGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockDurakGame) Reset() {
	_m.Called()
}

// GetGameEndFlag モック
func (_m *MockDurakGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsHumanTurn モック
func (_m *MockDurakGame) IsHumanTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// PlayerAttack モック
func (_m *MockDurakGame) PlayerAttack(cardIdx int) error {
	ret := _m.Called(cardIdx)
	return ret.Error(0)
}

// PlayerDefend モック
func (_m *MockDurakGame) PlayerDefend(attackIdx, handIdx int) error {
	ret := _m.Called(attackIdx, handIdx)
	return ret.Error(0)
}

// PlayerPass モック
func (_m *MockDurakGame) PlayerPass() error {
	ret := _m.Called()
	return ret.Error(0)
}

// PlayerTakeCards モック
func (_m *MockDurakGame) PlayerTakeCards() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CpuPlay モック
func (_m *MockDurakGame) CpuPlay() {
	_m.Called()
}

// HasPendingAction モック
func (_m *MockDurakGame) HasPendingAction() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// SetConfig モック
func (_m *MockDurakGame) SetConfig(config domain.DurakConfig) {
	_m.Called(config)
}

// SortHumanHand モック
func (_m *MockDurakGame) SortHumanHand(mode domain.DurakSortMode) error {
	ret := _m.Called(mode)
	return ret.Error(0)
}

// GetPlayerCnt モック
func (_m *MockDurakGame) GetPlayerCnt() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPlayer モック
func (_m *MockDurakGame) GetPlayer(i int) *domain.DurakPlayer {
	ret := _m.Called(i)
	if val, ok := ret.Get(0).(*domain.DurakPlayer); ok {
		return val
	}
	return nil
}

// GetCurrentTurn モック
func (_m *MockDurakGame) GetCurrentTurn() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetPhase モック
func (_m *MockDurakGame) GetPhase() domain.DurakPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.DurakPhase)
}

// GetAttackerIdx モック
func (_m *MockDurakGame) GetAttackerIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDefenderIdx モック
func (_m *MockDurakGame) GetDefenderIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTablePairs モック
func (_m *MockDurakGame) GetTablePairs() []*domain.DurakTablePair {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DurakTablePair); ok {
		return val
	}
	return nil
}

// GetTrumpSuit モック
func (_m *MockDurakGame) GetTrumpSuit() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetTrumpCard モック
func (_m *MockDurakGame) GetTrumpCard() *domain.Card {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.Card); ok {
		return val
	}
	return nil
}

// GetStockCount モック
func (_m *MockDurakGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetLoserIdx モック
func (_m *MockDurakGame) GetLoserIdx() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetConfig モック
func (_m *MockDurakGame) GetConfig() domain.DurakConfig {
	ret := _m.Called()
	if val, ok := ret.Get(0).(domain.DurakConfig); ok {
		return val
	}
	return domain.DurakConfig{}
}

// GetSortMode モック
func (_m *MockDurakGame) GetSortMode() domain.DurakSortMode {
	ret := _m.Called()
	return ret.Get(0).(domain.DurakSortMode)
}

// GetCpuActions モック
func (_m *MockDurakGame) GetCpuActions() []*domain.DurakCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.DurakCpuAction); ok {
		return val
	}
	return nil
}

// GetHumanAction モック
func (_m *MockDurakGame) GetHumanAction() *domain.DurakCpuAction {
	ret := _m.Called()
	if val, ok := ret.Get(0).(*domain.DurakCpuAction); ok {
		return val
	}
	return nil
}

// GetBoutNumber モック
func (_m *MockDurakGame) GetBoutNumber() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
// GetHint はサーバー計算の推奨手を返すモック。
func (_m *MockDurakGame) GetHint() *domain.DurakHint {
	out, _ := _m.Called().Get(0).(*domain.DurakHint)
	return out
}

func (_m *MockDurakGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	if val, ok := ret.Get(0).([]*domain.ActionLogEntry); ok {
		return val
	}
	return nil
}
