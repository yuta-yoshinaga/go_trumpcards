//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFortyThievesGame フォーティシーブスゲームモック
type MockFortyThievesGame struct {
	mock.Mock
}

func (_m *MockFortyThievesGame) Reset() {
	_m.Called()
}

func (_m *MockFortyThievesGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) GiveUp() {
	_m.Called()
}

func (_m *MockFortyThievesGame) GetHint() *domain.FortyThievesHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FortyThievesHint)
}

func (_m *MockFortyThievesGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortyThievesGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyThievesGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockFortyThievesGame) GetPhase() domain.FortyThievesPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FortyThievesPhase)
}

func (_m *MockFortyThievesGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyThievesGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyThievesGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockFortyThievesGame) GetTableau() [domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyThievesTableauCnt][]*domain.FortyThievesTableauCard)
}

func (_m *MockFortyThievesGame) GetFoundation() [domain.FortyThievesFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyThievesFoundationCnt][]*domain.Card)
}

func (_m *MockFortyThievesGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortyThievesGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockFortyThievesGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
