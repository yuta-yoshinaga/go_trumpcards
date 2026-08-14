//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRoyalCotillionGame ロイヤルコティヨン ゲームモック
type MockRoyalCotillionGame struct {
	mock.Mock
}

func (_m *MockRoyalCotillionGame) Reset() {
	_m.Called()
}

func (_m *MockRoyalCotillionGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) MoveReserveToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) MoveWasteToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) MoveStockToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) GiveUp() {
	_m.Called()
}

func (_m *MockRoyalCotillionGame) GetHint() *domain.RoyalCotillionHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.RoyalCotillionHint)
}

func (_m *MockRoyalCotillionGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRoyalCotillionGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRoyalCotillionGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockRoyalCotillionGame) GetPhase() domain.RoyalCotillionPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.RoyalCotillionPhase)
}

func (_m *MockRoyalCotillionGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRoyalCotillionGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRoyalCotillionGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockRoyalCotillionGame) GetTableau() [domain.RoyalCotillionTableauCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.RoyalCotillionTableauCnt]*domain.Card)
}

func (_m *MockRoyalCotillionGame) GetReserve() [domain.RoyalCotillionReserveCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.RoyalCotillionReserveCnt][]*domain.Card)
}

func (_m *MockRoyalCotillionGame) IsOddFoundation(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Get(0).(bool)
}

func (_m *MockRoyalCotillionGame) GetFoundation() [domain.RoyalCotillionFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.RoyalCotillionFoundationCnt][]*domain.Card)
}

func (_m *MockRoyalCotillionGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRoyalCotillionGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockRoyalCotillionGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockRoyalCotillionGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
