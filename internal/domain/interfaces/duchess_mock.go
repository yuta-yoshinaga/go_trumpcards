//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDuchessGame ダッチェス ゲームモック
type MockDuchessGame struct {
	mock.Mock
}

func (_m *MockDuchessGame) Reset() {
	_m.Called()
}

func (_m *MockDuchessGame) ChooseBaseRank(fanIdx int) error {
	ret := _m.Called(fanIdx)
	return ret.Error(0)
}

func (_m *MockDuchessGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveReserveToFoundation(fanIdx int) error {
	ret := _m.Called(fanIdx)
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveReserveToTableau(fanIdx, col int) error {
	ret := _m.Called(fanIdx, col)
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockDuchessGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockDuchessGame) GiveUp() {
	_m.Called()
}

func (_m *MockDuchessGame) GetHint() *domain.DuchessHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.DuchessHint)
}

func (_m *MockDuchessGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDuchessGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDuchessGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDuchessGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

// CanAutoComplete いま AutoComplete が 1 枚でも動かせるか
func (_m *MockDuchessGame) CanAutoComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDuchessGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockDuchessGame) GetPhase() domain.DuchessPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.DuchessPhase)
}

func (_m *MockDuchessGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDuchessGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDuchessGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockDuchessGame) GetReserve() [domain.DuchessReserveCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.DuchessReserveCnt][]*domain.Card)
}

func (_m *MockDuchessGame) GetTableau() [domain.DuchessTableauCnt][]*domain.DuchessTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.DuchessTableauCnt][]*domain.DuchessTableauCard)
}

func (_m *MockDuchessGame) GetFoundation() [domain.DuchessFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.DuchessFoundationCnt][]*domain.Card)
}

func (_m *MockDuchessGame) IsAwaitingBaseRank() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDuchessGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDuchessGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDuchessGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockDuchessGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockDuchessGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
