//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockScorpionGame スコーピオンゲームモック
type MockScorpionGame struct {
	mock.Mock
}

func (_m *MockScorpionGame) Reset() {
	_m.Called()
}

func (_m *MockScorpionGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockScorpionGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockScorpionGame) GiveUp() {
	_m.Called()
}

func (_m *MockScorpionGame) GetHint() *domain.ScorpionHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.ScorpionHint)
}

func (_m *MockScorpionGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockScorpionGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockScorpionGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockScorpionGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockScorpionGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockScorpionGame) GetPhase() domain.ScorpionPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.ScorpionPhase)
}

func (_m *MockScorpionGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockScorpionGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockScorpionGame) GetTableau() [domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.ScorpionTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockScorpionGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockScorpionGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockScorpionGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockScorpionGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
