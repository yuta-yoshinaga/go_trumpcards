//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockStalactitesGame フリーセルゲームモック
type MockStalactitesGame struct {
	mock.Mock
}

func (_m *MockStalactitesGame) Reset() {
	_m.Called()
}

func (_m *MockStalactitesGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) MoveTableauToStalactites(col, cell int) error {
	ret := _m.Called(col, cell)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) MoveStalactitesToTableau(cell, col int) error {
	ret := _m.Called(cell, col)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) MoveStalactitesToFoundation(cell int) error {
	ret := _m.Called(cell)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) GiveUp() {
	_m.Called()
}

func (_m *MockStalactitesGame) GetHint() *domain.StalactitesHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.StalactitesHint)
}

func (_m *MockStalactitesGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStalactitesGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStalactitesGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockStalactitesGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStalactitesGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockStalactitesGame) GetPhase() domain.StalactitesPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.StalactitesPhase)
}

func (_m *MockStalactitesGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStalactitesGame) GetTableau() [domain.StalactitesTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.StalactitesTableauCnt][]*domain.Card)
}

func (_m *MockStalactitesGame) GetCells() [domain.StalactitesCellCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.StalactitesCellCnt]*domain.Card)
}

func (_m *MockStalactitesGame) GetBaseRank() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockStalactitesGame) GetMaxMovableCards() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockStalactitesGame) GetMaxMovableCardsToEmptyColumn() int {
	args := _m.Called()
	return args.Int(0)
}

func (_m *MockStalactitesGame) GetFoundation() [domain.StalactitesFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.StalactitesFoundationCnt][]*domain.Card)
}

func (_m *MockStalactitesGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockStalactitesGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockStalactitesGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
