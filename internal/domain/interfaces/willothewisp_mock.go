//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWillOTheWispGame ウィル・オ・ザ・ウィスプゲームモック
type MockWillOTheWispGame struct {
	mock.Mock
}

// Reset mocks the Reset call.
func (_m *MockWillOTheWispGame) Reset() {
	_m.Called()
}

// Deal mocks the Deal call.
func (_m *MockWillOTheWispGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

// MoveTableauToTableau mocks the MoveTableauToTableau call.
func (_m *MockWillOTheWispGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

// GiveUp mocks the GiveUp call.
func (_m *MockWillOTheWispGame) GiveUp() {
	_m.Called()
}

// GetHint mocks the GetHint call.
func (_m *MockWillOTheWispGame) GetHint() *domain.WillOTheWispHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.WillOTheWispHint)
}

// AutoComplete mocks the AutoComplete call.
func (_m *MockWillOTheWispGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

// Undo mocks the Undo call.
func (_m *MockWillOTheWispGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo mocks the CanUndo call.
func (_m *MockWillOTheWispGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// UndoToEscape mocks the UndoToEscape call.
func (_m *MockWillOTheWispGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

// UndoN mocks the UndoN call.
func (_m *MockWillOTheWispGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

// GetPhase mocks the GetPhase call.
func (_m *MockWillOTheWispGame) GetPhase() domain.WillOTheWispPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.WillOTheWispPhase)
}

// GetMoveCount mocks the GetMoveCount call.
func (_m *MockWillOTheWispGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetStockCount mocks the GetStockCount call.
func (_m *MockWillOTheWispGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWillOTheWispGame) GetDealsRemaining() int {
	args := _m.Called()
	return args.Int(0)
}

// GetTableau mocks the GetTableau call.
func (_m *MockWillOTheWispGame) GetTableau() [domain.WillOTheWispTableauCnt][]*domain.WillOTheWispTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.WillOTheWispTableauCnt][]*domain.WillOTheWispTableauCard)
}

// GetCompletedSuits mocks the GetCompletedSuits call.
func (_m *MockWillOTheWispGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

// AllFaceUp mocks the AllFaceUp call.
func (_m *MockWillOTheWispGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetActionLog mocks the GetActionLog call.
func (_m *MockWillOTheWispGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetScore mocks the GetScore call.
func (_m *MockWillOTheWispGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsStalemate mocks the IsStalemate call.
func (_m *MockWillOTheWispGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockWillOTheWispGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
