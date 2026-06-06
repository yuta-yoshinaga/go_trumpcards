//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockEasthavenGame イーストヘイブンゲームモック
type MockEasthavenGame struct {
	mock.Mock
}

func (_m *MockEasthavenGame) Reset() {
	_m.Called()
}

func (_m *MockEasthavenGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockEasthavenGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockEasthavenGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockEasthavenGame) GiveUp() {
	_m.Called()
}

func (_m *MockEasthavenGame) GetHint() *domain.EasthavenHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.EasthavenHint)
}

func (_m *MockEasthavenGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockEasthavenGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockEasthavenGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockEasthavenGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockEasthavenGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockEasthavenGame) GetPhase() domain.EasthavenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.EasthavenPhase)
}

func (_m *MockEasthavenGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockEasthavenGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockEasthavenGame) GetTableau() [domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.EasthavenTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockEasthavenGame) GetFoundation() [domain.EasthavenFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.EasthavenFoundationCnt][]*domain.Card)
}

func (_m *MockEasthavenGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockEasthavenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockEasthavenGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockEasthavenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
