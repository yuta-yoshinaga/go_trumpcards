//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWaspGame ワスプゲームモック
type MockWaspGame struct {
	mock.Mock
}

func (_m *MockWaspGame) Reset() {
	_m.Called()
}

func (_m *MockWaspGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWaspGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockWaspGame) GiveUp() {
	_m.Called()
}

func (_m *MockWaspGame) GetHint() *domain.WaspHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.WaspHint)
}

func (_m *MockWaspGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWaspGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWaspGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWaspGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWaspGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockWaspGame) GetPhase() domain.WaspPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.WaspPhase)
}

func (_m *MockWaspGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWaspGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWaspGame) GetTableau() [domain.WaspTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.WaspTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockWaspGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWaspGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWaspGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWaspGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockWaspGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
