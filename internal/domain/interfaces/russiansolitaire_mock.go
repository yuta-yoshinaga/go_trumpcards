//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRussianSolitaireGame ロシアンソリティアゲームモック
type MockRussianSolitaireGame struct {
	mock.Mock
}

func (_m *MockRussianSolitaireGame) Reset() {
	_m.Called()
}

func (_m *MockRussianSolitaireGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockRussianSolitaireGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockRussianSolitaireGame) GiveUp() {
	_m.Called()
}

func (_m *MockRussianSolitaireGame) GetHint() *domain.RussianSolitaireHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.RussianSolitaireHint)
}

func (_m *MockRussianSolitaireGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRussianSolitaireGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRussianSolitaireGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRussianSolitaireGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRussianSolitaireGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockRussianSolitaireGame) GetPhase() domain.RussianSolitairePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.RussianSolitairePhase)
}

func (_m *MockRussianSolitaireGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRussianSolitaireGame) GetTableau() [domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.RussianSolitaireTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockRussianSolitaireGame) GetFoundation() [domain.RussianSolitaireFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.RussianSolitaireFoundationCnt][]*domain.Card)
}

func (_m *MockRussianSolitaireGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRussianSolitaireGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRussianSolitaireGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockRussianSolitaireGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
