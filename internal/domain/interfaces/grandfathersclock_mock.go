//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGrandfathersClockGame グランドファーザーズ・クロック ゲームモック
type MockGrandfathersClockGame struct {
	mock.Mock
}

func (_m *MockGrandfathersClockGame) Reset() {
	_m.Called()
}

func (_m *MockGrandfathersClockGame) MoveTableauToFoundation(col, fIdx int) error {
	ret := _m.Called(col, fIdx)
	return ret.Error(0)
}

func (_m *MockGrandfathersClockGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockGrandfathersClockGame) GiveUp() {
	_m.Called()
}

func (_m *MockGrandfathersClockGame) GetHint() *domain.GrandfathersClockHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.GrandfathersClockHint)
}

func (_m *MockGrandfathersClockGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGrandfathersClockGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGrandfathersClockGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGrandfathersClockGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGrandfathersClockGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockGrandfathersClockGame) GetPhase() domain.GrandfathersClockPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GrandfathersClockPhase)
}

func (_m *MockGrandfathersClockGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGrandfathersClockGame) GetFoundation() [domain.GrandfathersClockFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.GrandfathersClockFoundationCnt][]*domain.Card)
}

func (_m *MockGrandfathersClockGame) GetTableau() [domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.GrandfathersClockTableauCnt][]*domain.GrandfathersClockTableauCard)
}

func (_m *MockGrandfathersClockGame) IsFoundationComplete(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Bool(0)
}

func (_m *MockGrandfathersClockGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGrandfathersClockGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockGrandfathersClockGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockGrandfathersClockGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
