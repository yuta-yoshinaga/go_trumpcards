//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAgnesGame アグネス・ソレルゲームモック
type MockAgnesGame struct {
	mock.Mock
}

func (_m *MockAgnesGame) Reset() { _m.Called() }

func (_m *MockAgnesGame) DealStock() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAgnesGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockAgnesGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAgnesGame) GiveUp() { _m.Called() }

func (_m *MockAgnesGame) GetHint() *domain.AgnesHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.AgnesHint)
}

func (_m *MockAgnesGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAgnesGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAgnesGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockAgnesGame) GetPhase() domain.AgnesPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AgnesPhase)
}

func (_m *MockAgnesGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAgnesGame) IsStalemate() bool { return _m.Called().Bool(0) }

func (_m *MockAgnesGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAgnesGame) GetTableau() [domain.AgnesTableauCnt][]*domain.AgnesTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.AgnesTableauCnt][]*domain.AgnesTableauCard)
}

func (_m *MockAgnesGame) GetFoundation() [domain.AgnesFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.AgnesFoundationCnt][]*domain.Card)
}

func (_m *MockAgnesGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAgnesGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAgnesGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
