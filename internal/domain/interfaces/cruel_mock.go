//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCruelGame クルーエルゲームモック
type MockCruelGame struct {
	mock.Mock
}

func (_m *MockCruelGame) Reset() {
	_m.Called()
}

func (_m *MockCruelGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockCruelGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockCruelGame) Shift() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCruelGame) GiveUp() {
	_m.Called()
}

func (_m *MockCruelGame) GetHint() *domain.CruelHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CruelHint)
}

func (_m *MockCruelGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCruelGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCruelGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// CanAutoComplete モック
func (_m *MockCruelGame) CanAutoComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCruelGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCruelGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCruelGame) GetPhase() domain.CruelPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CruelPhase)
}

func (_m *MockCruelGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCruelGame) GetTableau() [domain.CruelTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.CruelTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockCruelGame) GetFoundation() [domain.CruelFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CruelFoundationCnt][]*domain.Card)
}

func (_m *MockCruelGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCruelGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCruelGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
