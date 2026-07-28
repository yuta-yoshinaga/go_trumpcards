//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBisleyGame ビズリー ゲームモック
type MockBisleyGame struct {
	mock.Mock
}

func (_m *MockBisleyGame) Reset() {
	_m.Called()
}

func (_m *MockBisleyGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockBisleyGame) MoveTableauToAceFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockBisleyGame) MoveTableauToKingFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockBisleyGame) GiveUp() {
	_m.Called()
}

func (_m *MockBisleyGame) GetHint() *domain.BisleyHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BisleyHint)
}

func (_m *MockBisleyGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBisleyGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBisleyGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBisleyGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBisleyGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBisleyGame) GetPhase() domain.BisleyPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BisleyPhase)
}

func (_m *MockBisleyGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBisleyGame) GetTableau() [domain.BisleyTableauCnt][]*domain.BisleyTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.BisleyTableauCnt][]*domain.BisleyTableauCard)
}

func (_m *MockBisleyGame) GetAceFoundations() [domain.BisleyFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BisleyFoundationCnt][]*domain.Card)
}

func (_m *MockBisleyGame) GetKingFoundations() [domain.BisleyFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BisleyFoundationCnt][]*domain.Card)
}

func (_m *MockBisleyGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBisleyGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockBisleyGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBisleyGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
