//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMonteCarloGame はモンテカルロ・ソリティアのモック。
type MockMonteCarloGame struct {
	mock.Mock
}

// Reset モック
// CountRemovablePairs は取り除ける組の数を返す。
func (m *MockMonteCarloGame) CountRemovablePairs() int {
	args := m.Called()
	return args.Int(0)
}

func (_m *MockMonteCarloGame) Reset() { _m.Called() }

// Remove モック
func (_m *MockMonteCarloGame) Remove(r1, c1, r2, c2 int) error {
	ret := _m.Called(r1, c1, r2, c2)
	return ret.Error(0)
}

// Deal モック
func (_m *MockMonteCarloGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Undo モック
func (_m *MockMonteCarloGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo モック
func (_m *MockMonteCarloGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GiveUp モック
func (_m *MockMonteCarloGame) GiveUp() { _m.Called() }

// Hint モック
func (_m *MockMonteCarloGame) Hint() *domain.MonteCarloHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.MonteCarloHint)
}

// GetPhase モック
func (_m *MockMonteCarloGame) GetPhase() domain.MonteCarloPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MonteCarloPhase)
}

// GetBoard モック
func (_m *MockMonteCarloGame) GetBoard() [domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.MonteCarloGridSize][domain.MonteCarloGridSize]*domain.Card)
}

// GetStockCount モック
func (_m *MockMonteCarloGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetRemovedCount モック
func (_m *MockMonteCarloGame) GetRemovedCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetDealCount モック
func (_m *MockMonteCarloGame) GetDealCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsComplete モック
func (_m *MockMonteCarloGame) IsComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsStalemate モック
func (_m *MockMonteCarloGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActionLog モック
func (_m *MockMonteCarloGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag モック
func (_m *MockMonteCarloGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
