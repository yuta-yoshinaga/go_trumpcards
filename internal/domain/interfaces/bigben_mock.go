//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBigBenGame ビッグ・ベン ゲームモック
type MockBigBenGame struct {
	mock.Mock
}

func (_m *MockBigBenGame) Reset() {
	_m.Called()
}

func (_m *MockBigBenGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBigBenGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBigBenGame) MoveTableauToFoundation(col, fIdx int) error {
	ret := _m.Called(col, fIdx)
	return ret.Error(0)
}

func (_m *MockBigBenGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockBigBenGame) GiveUp() {
	_m.Called()
}

func (_m *MockBigBenGame) GetHint() *domain.BigBenHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BigBenHint)
}

func (_m *MockBigBenGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBigBenGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBigBenGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBigBenGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBigBenGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBigBenGame) GetPhase() domain.BigBenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BigBenPhase)
}

func (_m *MockBigBenGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBigBenGame) GetFoundation() [domain.BigBenFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BigBenFoundationCnt][]*domain.Card)
}

func (_m *MockBigBenGame) GetTableau() [domain.BigBenTableauCnt][]*domain.BigBenTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.BigBenTableauCnt][]*domain.BigBenTableauCard)
}

func (_m *MockBigBenGame) IsFoundationComplete(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Bool(0)
}

func (_m *MockBigBenGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBigBenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockBigBenGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBigBenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
