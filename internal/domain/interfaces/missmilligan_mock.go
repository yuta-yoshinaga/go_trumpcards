//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMissMilliganGame ミス・ミリガン ゲームモック
type MockMissMilliganGame struct {
	mock.Mock
}

func (_m *MockMissMilliganGame) Reset() {
	_m.Called()
}

func (_m *MockMissMilliganGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) Waive(col, cardIndex int) error {
	ret := _m.Called(col, cardIndex)
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) PlaceWaived(toCol int) error {
	ret := _m.Called(toCol)
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) MoveWaivedToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) GiveUp() {
	_m.Called()
}

func (_m *MockMissMilliganGame) GetHint() *domain.MissMilliganHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.MissMilliganHint)
}

func (_m *MockMissMilliganGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMissMilliganGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMissMilliganGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockMissMilliganGame) GetPhase() domain.MissMilliganPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MissMilliganPhase)
}

func (_m *MockMissMilliganGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMissMilliganGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMissMilliganGame) GetWaived() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockMissMilliganGame) CanWaive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMissMilliganGame) GetTableau() [domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.MissMilliganTableauCnt][]*domain.MissMilliganTableauCard)
}

func (_m *MockMissMilliganGame) GetFoundation() [domain.MissMilliganFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.MissMilliganFoundationCnt][]*domain.Card)
}

func (_m *MockMissMilliganGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMissMilliganGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockMissMilliganGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockMissMilliganGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
