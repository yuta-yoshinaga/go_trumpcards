//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSomersetGame Somerset ゲームモック
type MockSomersetGame struct {
	mock.Mock
}

func (_m *MockSomersetGame) Reset() {
	_m.Called()
}

func (_m *MockSomersetGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockSomersetGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockSomersetGame) GiveUp() {
	_m.Called()
}

func (_m *MockSomersetGame) GetHint() *domain.SomersetHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SomersetHint)
}

func (_m *MockSomersetGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSomersetGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSomersetGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSomersetGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSomersetGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSomersetGame) GetPhase() domain.SomersetPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SomersetPhase)
}

func (_m *MockSomersetGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSomersetGame) GetTableau() [domain.SomersetTableauCnt][]*domain.SomersetTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.SomersetTableauCnt][]*domain.SomersetTableauCard)
}

func (_m *MockSomersetGame) GetFoundation() [domain.SomersetFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SomersetFoundationCnt][]*domain.Card)
}

func (_m *MockSomersetGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSomersetGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSomersetGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSomersetGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
