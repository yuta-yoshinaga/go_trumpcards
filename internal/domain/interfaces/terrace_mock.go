//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTerraceGame テラス ゲームモック
type MockTerraceGame struct {
	mock.Mock
}

func (_m *MockTerraceGame) Reset() {
	_m.Called()
}

func (_m *MockTerraceGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTerraceGame) MoveReserveToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTerraceGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTerraceGame) MoveWasteToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockTerraceGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockTerraceGame) MoveTableauToTableau(fromPile, toPile int) error {
	ret := _m.Called(fromPile, toPile)
	return ret.Error(0)
}

func (_m *MockTerraceGame) GiveUp() {
	_m.Called()
}

func (_m *MockTerraceGame) GetHint() *domain.TerraceHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.TerraceHint)
}

func (_m *MockTerraceGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTerraceGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTerraceGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockTerraceGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTerraceGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockTerraceGame) GetPhase() domain.TerracePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TerracePhase)
}

func (_m *MockTerraceGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTerraceGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTerraceGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockTerraceGame) GetReserve() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockTerraceGame) GetTableau() [domain.TerraceTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.TerraceTableauCnt][]*domain.Card)
}

func (_m *MockTerraceGame) GetFoundation() [domain.TerraceFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.TerraceFoundationCnt][]*domain.Card)
}

func (_m *MockTerraceGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTerraceGame) IsAwaitingBaseRank() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockTerraceGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockTerraceGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockTerraceGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockTerraceGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
