//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrescentGame クレセント・ソリティアのモック。
type MockCrescentGame struct {
	mock.Mock
}

func (_m *MockCrescentGame) Reset() {
	_m.Called()
}

func (_m *MockCrescentGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockCrescentGame) MoveTableauToFoundation(fromCol, foundationIdx int) error {
	ret := _m.Called(fromCol, foundationIdx)
	return ret.Error(0)
}

func (_m *MockCrescentGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrescentGame) GiveUp() {
	_m.Called()
}

func (_m *MockCrescentGame) GetHint() *domain.CrescentHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CrescentHint)
}

func (_m *MockCrescentGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrescentGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrescentGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCrescentGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrescentGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCrescentGame) GetPhase() domain.CrescentPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CrescentPhase)
}

func (_m *MockCrescentGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrescentGame) GetRedealsRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrescentGame) GetTableau() [domain.CrescentTableauCnt][]*domain.CrescentTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.CrescentTableauCnt][]*domain.CrescentTableauCard)
}

func (_m *MockCrescentGame) GetFoundation() [domain.CrescentFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CrescentFoundationCnt][]*domain.Card)
}

func (_m *MockCrescentGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockCrescentGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCrescentGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
