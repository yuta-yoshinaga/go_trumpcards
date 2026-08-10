//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockColoradoGame コロラド ゲームモック
type MockColoradoGame struct {
	mock.Mock
}

func (_m *MockColoradoGame) Reset() {
	_m.Called()
}

func (_m *MockColoradoGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockColoradoGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockColoradoGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockColoradoGame) MoveWasteToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockColoradoGame) MoveStockToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockColoradoGame) GiveUp() {
	_m.Called()
}

func (_m *MockColoradoGame) GetHint() *domain.ColoradoHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ColoradoHint)
}

func (_m *MockColoradoGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockColoradoGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockColoradoGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockColoradoGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockColoradoGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockColoradoGame) GetPhase() domain.ColoradoPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ColoradoPhase)
}

func (_m *MockColoradoGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockColoradoGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockColoradoGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockColoradoGame) GetTableau() [domain.ColoradoTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.ColoradoTableauCnt][]*domain.Card)
}

func (_m *MockColoradoGame) GetFoundation() [domain.ColoradoFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.ColoradoFoundationCnt][]*domain.Card)
}

func (_m *MockColoradoGame) IsAscendingFoundation(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Get(0).(bool)
}

func (_m *MockColoradoGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockColoradoGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockColoradoGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockColoradoGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
