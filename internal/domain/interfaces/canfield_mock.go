//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCanfieldGame キャンフィールドゲームモック
type MockCanfieldGame struct {
	mock.Mock
}

func (_m *MockCanfieldGame) Reset() { _m.Called() }

func (_m *MockCanfieldGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveReserveToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockCanfieldGame) MoveReserveToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCanfieldGame) GiveUp() { _m.Called() }

func (_m *MockCanfieldGame) GetHint() *domain.CanfieldHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CanfieldHint)
}

func (_m *MockCanfieldGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCanfieldGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCanfieldGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCanfieldGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCanfieldGame) GetPhase() domain.CanfieldPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CanfieldPhase)
}

func (_m *MockCanfieldGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCanfieldGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCanfieldGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockCanfieldGame) GetReserve() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockCanfieldGame) GetTableau() [domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.CanfieldTableauCnt][]*domain.CanfieldTableauCard)
}

func (_m *MockCanfieldGame) GetFoundation() [domain.CanfieldFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CanfieldFoundationCnt][]*domain.Card)
}

func (_m *MockCanfieldGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCanfieldGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCanfieldGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
