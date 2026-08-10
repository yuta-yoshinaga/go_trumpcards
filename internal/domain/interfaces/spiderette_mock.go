//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpideretteGame スパイダレットゲームモック
type MockSpideretteGame struct {
	mock.Mock
}

// Reset mocks the Reset call.
func (_m *MockSpideretteGame) Reset() {
	_m.Called()
}

// Deal mocks the Deal call.
func (_m *MockSpideretteGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

// MoveTableauToTableau mocks the MoveTableauToTableau call.
func (_m *MockSpideretteGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

// GiveUp mocks the GiveUp call.
func (_m *MockSpideretteGame) GiveUp() {
	_m.Called()
}

// GetHint mocks the GetHint call.
func (_m *MockSpideretteGame) GetHint() *domain.SpideretteHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SpideretteHint)
}

// AutoComplete mocks the AutoComplete call.
func (_m *MockSpideretteGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Undo mocks the Undo call.
func (_m *MockSpideretteGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo mocks the CanUndo call.
func (_m *MockSpideretteGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// UndoToEscape mocks the UndoToEscape call.
func (_m *MockSpideretteGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

// UndoN mocks the UndoN call.
func (_m *MockSpideretteGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

// GetPhase mocks the GetPhase call.
func (_m *MockSpideretteGame) GetPhase() domain.SpiderettePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SpiderettePhase)
}

// GetMoveCount mocks the GetMoveCount call.
func (_m *MockSpideretteGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetStockCount mocks the GetStockCount call.
func (_m *MockSpideretteGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpideretteGame) GetDealsRemaining() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTableau mocks the GetTableau call.
func (_m *MockSpideretteGame) GetTableau() [domain.SpideretteTableauCnt][]*domain.SpideretteTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.SpideretteTableauCnt][]*domain.SpideretteTableauCard)
}

// GetCompletedSuits mocks the GetCompletedSuits call.
func (_m *MockSpideretteGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

// AllFaceUp mocks the AllFaceUp call.
func (_m *MockSpideretteGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActionLog mocks the GetActionLog call.
func (_m *MockSpideretteGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetScore mocks the GetScore call.
func (_m *MockSpideretteGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsStalemate mocks the IsStalemate call.
func (_m *MockSpideretteGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSpideretteGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
