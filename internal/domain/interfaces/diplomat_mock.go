//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDiplomatGame ディプロマット ゲームモック
type MockDiplomatGame struct {
	mock.Mock
}

func (_m *MockDiplomatGame) Reset() {
	_m.Called()
}

func (_m *MockDiplomatGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDiplomatGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockDiplomatGame) MoveTableauToTableau(fromPile, toPile int) error {
	ret := _m.Called(fromPile, toPile)
	return ret.Error(0)
}

func (_m *MockDiplomatGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDiplomatGame) MoveWasteToTableau(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockDiplomatGame) GiveUp() {
	_m.Called()
}

func (_m *MockDiplomatGame) GetHint() *domain.DiplomatHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.DiplomatHint)
}

func (_m *MockDiplomatGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDiplomatGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockDiplomatGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDiplomatGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDiplomatGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockDiplomatGame) GetPhase() domain.DiplomatPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.DiplomatPhase)
}

func (_m *MockDiplomatGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDiplomatGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockDiplomatGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockDiplomatGame) GetTableau() [domain.DiplomatTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.DiplomatTableauCnt][]*domain.Card)
}

func (_m *MockDiplomatGame) GetFoundation() [domain.DiplomatFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.DiplomatFoundationCnt][]*domain.Card)
}

func (_m *MockDiplomatGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockDiplomatGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockDiplomatGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockDiplomatGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
