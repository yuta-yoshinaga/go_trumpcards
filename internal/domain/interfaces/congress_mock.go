//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCongressGame コングレス ゲームモック
type MockCongressGame struct {
	mock.Mock
}

func (_m *MockCongressGame) Reset() {
	_m.Called()
}

func (_m *MockCongressGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCongressGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockCongressGame) MoveTableauToTableau(fromPile, toPile int) error {
	ret := _m.Called(fromPile, toPile)
	return ret.Error(0)
}

func (_m *MockCongressGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCongressGame) MoveWasteToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockCongressGame) MoveStockToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockCongressGame) GiveUp() {
	_m.Called()
}

func (_m *MockCongressGame) GetHint() *domain.CongressHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CongressHint)
}

func (_m *MockCongressGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCongressGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCongressGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCongressGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCongressGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCongressGame) GetPhase() domain.CongressPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CongressPhase)
}

func (_m *MockCongressGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCongressGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCongressGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockCongressGame) GetTableau() [domain.CongressTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CongressTableauCnt][]*domain.Card)
}

func (_m *MockCongressGame) GetFoundation() [domain.CongressFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CongressFoundationCnt][]*domain.Card)
}

func (_m *MockCongressGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCongressGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockCongressGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCongressGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
