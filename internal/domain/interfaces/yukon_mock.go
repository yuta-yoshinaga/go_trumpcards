//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockYukonGame ユーコンゲームモック
type MockYukonGame struct {
	mock.Mock
}

func (_m *MockYukonGame) Reset() {
	_m.Called()
}

func (_m *MockYukonGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockYukonGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockYukonGame) GiveUp() {
	_m.Called()
}

func (_m *MockYukonGame) GetHint() *domain.YukonHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.YukonHint)
}

func (_m *MockYukonGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockYukonGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockYukonGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockYukonGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockYukonGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockYukonGame) GetPhase() domain.YukonPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.YukonPhase)
}

func (_m *MockYukonGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockYukonGame) GetTableau() [domain.YukonTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.YukonTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockYukonGame) GetFoundation() [domain.YukonFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.YukonFoundationCnt][]*domain.Card)
}

func (_m *MockYukonGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockYukonGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockYukonGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
