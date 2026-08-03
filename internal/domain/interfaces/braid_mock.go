//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBraidGame ブレイド ゲームモック
type MockBraidGame struct {
	mock.Mock
}

func (_m *MockBraidGame) Reset() {
	_m.Called()
}

func (_m *MockBraidGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBraidGame) ChooseDirection(ascending bool) error {
	ret := _m.Called(ascending)
	return ret.Error(0)
}

func (_m *MockBraidGame) MoveBraidToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBraidGame) MoveFieldToFoundation(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockBraidGame) MoveHelperToFoundation(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockBraidGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBraidGame) MoveWasteToHelper(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockBraidGame) GiveUp() {
	_m.Called()
}

func (_m *MockBraidGame) GetHint() *domain.BraidHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BraidHint)
}

func (_m *MockBraidGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBraidGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBraidGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBraidGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBraidGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBraidGame) GetPhase() domain.BraidPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BraidPhase)
}

func (_m *MockBraidGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBraidGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBraidGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockBraidGame) GetBraid() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockBraidGame) GetFields() [domain.BraidFieldCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BraidFieldCnt]*domain.Card)
}

func (_m *MockBraidGame) GetHelpers() [domain.BraidHelperCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BraidHelperCnt]*domain.Card)
}

func (_m *MockBraidGame) GetFoundation() [domain.BraidFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BraidFoundationCnt][]*domain.Card)
}

func (_m *MockBraidGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBraidGame) GetDirection() domain.BraidDirection {
	ret := _m.Called()
	return ret.Get(0).(domain.BraidDirection)
}

func (_m *MockBraidGame) IsAwaitingDirection() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBraidGame) GetPassesUsed() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBraidGame) CanRedeal() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBraidGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBraidGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockBraidGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBraidGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
