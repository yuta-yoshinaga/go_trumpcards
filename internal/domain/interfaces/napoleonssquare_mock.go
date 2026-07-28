//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockNapoleonsSquareGame ナポレオンズ・スクエア ゲームモック
type MockNapoleonsSquareGame struct {
	mock.Mock
}

func (_m *MockNapoleonsSquareGame) Reset() {
	_m.Called()
}

func (_m *MockNapoleonsSquareGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) GiveUp() {
	_m.Called()
}

func (_m *MockNapoleonsSquareGame) GetHint() *domain.NapoleonsSquareHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.NapoleonsSquareHint)
}

func (_m *MockNapoleonsSquareGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockNapoleonsSquareGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNapoleonsSquareGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockNapoleonsSquareGame) GetPhase() domain.NapoleonsSquarePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.NapoleonsSquarePhase)
}

func (_m *MockNapoleonsSquareGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNapoleonsSquareGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockNapoleonsSquareGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockNapoleonsSquareGame) GetTableau() [domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.NapoleonsSquareTableauCnt][]*domain.NapoleonsSquareTableauCard)
}

func (_m *MockNapoleonsSquareGame) GetFoundation() [domain.NapoleonsSquareFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.NapoleonsSquareFoundationCnt][]*domain.Card)
}

func (_m *MockNapoleonsSquareGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockNapoleonsSquareGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockNapoleonsSquareGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockNapoleonsSquareGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
