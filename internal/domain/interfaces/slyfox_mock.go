//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSlyFoxGame スライ・フォックス ゲームモック
type MockSlyFoxGame struct {
	mock.Mock
}

func (_m *MockSlyFoxGame) Reset() {
	_m.Called()
}

func (_m *MockSlyFoxGame) DealToPile(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) DealToFoundation(fIdx int) error {
	ret := _m.Called(fIdx)
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) DealtThisCycle() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSlyFoxGame) ReserveIsLocked() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSlyFoxGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) GiveUp() {
	_m.Called()
}

func (_m *MockSlyFoxGame) GetHint() *domain.SlyFoxHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SlyFoxHint)
}

func (_m *MockSlyFoxGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSlyFoxGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSlyFoxGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSlyFoxGame) GetPhase() domain.SlyFoxPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SlyFoxPhase)
}

func (_m *MockSlyFoxGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSlyFoxGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSlyFoxGame) GetTableau() [domain.SlyFoxTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SlyFoxTableauCnt][]*domain.Card)
}

func (_m *MockSlyFoxGame) GetFoundation() [domain.SlyFoxFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SlyFoxFoundationCnt][]*domain.Card)
}

func (_m *MockSlyFoxGame) IsAscendingFoundation(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Get(0).(bool)
}

func (_m *MockSlyFoxGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSlyFoxGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSlyFoxGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSlyFoxGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
