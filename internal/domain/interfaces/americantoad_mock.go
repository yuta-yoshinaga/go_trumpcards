//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAmericanToadGame アメリカン・トード ゲームモック
type MockAmericanToadGame struct {
	mock.Mock
}

func (_m *MockAmericanToadGame) Reset() {
	_m.Called()
}

func (_m *MockAmericanToadGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveReserveToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveReserveToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) GiveUp() {
	_m.Called()
}

func (_m *MockAmericanToadGame) GetHint() *domain.AmericanToadHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.AmericanToadHint)
}

func (_m *MockAmericanToadGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAmericanToadGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAmericanToadGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockAmericanToadGame) GetPhase() domain.AmericanToadPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AmericanToadPhase)
}

func (_m *MockAmericanToadGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAmericanToadGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAmericanToadGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockAmericanToadGame) GetReserve() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockAmericanToadGame) GetTableau() [domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.AmericanToadTableauCnt][]*domain.AmericanToadTableauCard)
}

func (_m *MockAmericanToadGame) GetFoundation() [domain.AmericanToadFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.AmericanToadFoundationCnt][]*domain.Card)
}

func (_m *MockAmericanToadGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAmericanToadGame) GetPassesUsed() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAmericanToadGame) CanRedeal() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAmericanToadGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAmericanToadGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockAmericanToadGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAmericanToadGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
