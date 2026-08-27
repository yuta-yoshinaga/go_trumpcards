//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSeahavenTowersGame シーヘイブンタワーズゲームモック
type MockSeahavenTowersGame struct {
	mock.Mock
}

func (_m *MockSeahavenTowersGame) Reset() {
	_m.Called()
}

func (_m *MockSeahavenTowersGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) MoveTableauToFreeCell(col, cell int) error {
	ret := _m.Called(col, cell)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) MoveFreeCellToTableau(cell, col int) error {
	ret := _m.Called(cell, col)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) MoveFreeCellToFoundation(cell int) error {
	ret := _m.Called(cell)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) GiveUp() {
	_m.Called()
}

func (_m *MockSeahavenTowersGame) GetHint() *domain.SeahavenTowersHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SeahavenTowersHint)
}

func (_m *MockSeahavenTowersGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSeahavenTowersGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSeahavenTowersGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSeahavenTowersGame) GetPhase() domain.SeahavenTowersPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SeahavenTowersPhase)
}

func (_m *MockSeahavenTowersGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSeahavenTowersGame) GetTableau() [domain.SeahavenTowersTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SeahavenTowersTableauCnt][]*domain.Card)
}

func (_m *MockSeahavenTowersGame) GetFreeCells() [domain.SeahavenTowersCellCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SeahavenTowersCellCnt]*domain.Card)
}

func (_m *MockSeahavenTowersGame) GetFoundation() [domain.SeahavenTowersFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SeahavenTowersFoundationCnt][]*domain.Card)
}

func (_m *MockSeahavenTowersGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSeahavenTowersGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSeahavenTowersGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// CanAutoComplete mocks the CanAutoComplete call.
func (_m *MockSeahavenTowersGame) CanAutoComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
