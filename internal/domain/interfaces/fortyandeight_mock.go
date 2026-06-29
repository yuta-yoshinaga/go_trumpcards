//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFortyAndEightGame フォーティ・アンド・エイトゲームモック
type MockFortyAndEightGame struct {
	mock.Mock
}

func (_m *MockFortyAndEightGame) Reset() {
	_m.Called()
}

func (_m *MockFortyAndEightGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) GiveUp() {
	_m.Called()
}

func (_m *MockFortyAndEightGame) GetHint() *domain.FortyAndEightHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FortyAndEightHint)
}

func (_m *MockFortyAndEightGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortyAndEightGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyAndEightGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockFortyAndEightGame) GetPhase() domain.FortyAndEightPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.FortyAndEightPhase)
}

func (_m *MockFortyAndEightGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyAndEightGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockFortyAndEightGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockFortyAndEightGame) GetTableau() [domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyAndEightTableauCnt][]*domain.FortyAndEightTableauCard)
}

func (_m *MockFortyAndEightGame) GetFoundation() [domain.FortyAndEightFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.FortyAndEightFoundationCnt][]*domain.Card)
}

func (_m *MockFortyAndEightGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockFortyAndEightGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockFortyAndEightGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRedealUsed mocks the GetRedealUsed call.
func (_m *MockFortyAndEightGame) GetRedealUsed() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// CanRedeal mocks the CanRedeal call.
func (_m *MockFortyAndEightGame) CanRedeal() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockFortyAndEightGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
