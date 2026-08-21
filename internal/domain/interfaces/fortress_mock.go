//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFortressGame Fortress ゲームモック
type MockFortressGame struct {
	mock.Mock
}

func (_m *MockFortressGame) Reset() {
	_m.Called()
}

func (_m *MockFortressGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockFortressGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFortressGame) GiveUp() {
	_m.Called()
}

func (_m *MockFortressGame) GetHint() *domain.FortressHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FortressHint)
}

func (_m *MockFortressGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortressGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortressGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortressGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortressGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockFortressGame) GetPhase() domain.FortressPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FortressPhase)
}

func (_m *MockFortressGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortressGame) GetTableau() [domain.FortressTableauCnt][]*domain.FortressTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.FortressTableauCnt][]*domain.FortressTableauCard)
}

func (_m *MockFortressGame) GetFoundation() [domain.FortressFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FortressFoundationCnt][]*domain.Card)
}

func (_m *MockFortressGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortressGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockFortressGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockFortressGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
