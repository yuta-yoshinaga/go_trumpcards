//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBeleagueredCastleGame Beleaguered Castle ゲームモック
type MockBeleagueredCastleGame struct {
	mock.Mock
}

func (_m *MockBeleagueredCastleGame) Reset() {
	_m.Called()
}

func (_m *MockBeleagueredCastleGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockBeleagueredCastleGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockBeleagueredCastleGame) GiveUp() {
	_m.Called()
}

func (_m *MockBeleagueredCastleGame) GetHint() *domain.BeleagueredCastleHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BeleagueredCastleHint)
}

func (_m *MockBeleagueredCastleGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBeleagueredCastleGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBeleagueredCastleGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBeleagueredCastleGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBeleagueredCastleGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBeleagueredCastleGame) GetPhase() domain.BeleagueredCastlePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BeleagueredCastlePhase)
}

func (_m *MockBeleagueredCastleGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBeleagueredCastleGame) GetTableau() [domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.BeleagueredCastleTableauCnt][]*domain.BeleagueredCastleTableauCard)
}

func (_m *MockBeleagueredCastleGame) GetFoundation() [domain.BeleagueredCastleFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BeleagueredCastleFoundationCnt][]*domain.Card)
}

func (_m *MockBeleagueredCastleGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBeleagueredCastleGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockBeleagueredCastleGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBeleagueredCastleGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
