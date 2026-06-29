//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFlowerGardenGame Flower Garden ゲームモック
type MockFlowerGardenGame struct {
	mock.Mock
}

func (_m *MockFlowerGardenGame) Reset() {
	_m.Called()
}

func (_m *MockFlowerGardenGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) MoveReserveToTableau(reserveIdx, toCol int) error {
	ret := _m.Called(reserveIdx, toCol)
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) MoveReserveToFoundation(reserveIdx int) error {
	ret := _m.Called(reserveIdx)
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) GiveUp() {
	_m.Called()
}

func (_m *MockFlowerGardenGame) GetHint() *domain.FlowerGardenHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FlowerGardenHint)
}

func (_m *MockFlowerGardenGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFlowerGardenGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFlowerGardenGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockFlowerGardenGame) GetPhase() domain.FlowerGardenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FlowerGardenPhase)
}

func (_m *MockFlowerGardenGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFlowerGardenGame) GetTableau() [domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.FlowerGardenTableauCnt][]*domain.FlowerGardenTableauCard)
}

func (_m *MockFlowerGardenGame) GetReserve() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockFlowerGardenGame) GetFoundation() [domain.FlowerGardenFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FlowerGardenFoundationCnt][]*domain.Card)
}

func (_m *MockFlowerGardenGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFlowerGardenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockFlowerGardenGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockFlowerGardenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
