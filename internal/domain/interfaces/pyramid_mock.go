//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPyramidGame ピラミッドゲームモック
type MockPyramidGame struct {
	mock.Mock
}

func (_m *MockPyramidGame) Reset() {
	_m.Called()
}

func (_m *MockPyramidGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPyramidGame) RemovePair(row1, col1, row2, col2 int) error {
	ret := _m.Called(row1, col1, row2, col2)
	return ret.Error(0)
}

func (_m *MockPyramidGame) RemoveKing(row, col int) error {
	ret := _m.Called(row, col)
	return ret.Error(0)
}

func (_m *MockPyramidGame) RemoveWithWaste(row, col int) error {
	ret := _m.Called(row, col)
	return ret.Error(0)
}

func (_m *MockPyramidGame) RemoveWasteKing() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPyramidGame) GiveUp() {
	_m.Called()
}

func (_m *MockPyramidGame) GetHint() *domain.PyramidHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.PyramidHint)
}

func (_m *MockPyramidGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPyramidGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPyramidGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPyramidGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockPyramidGame) GetPhase() domain.PyramidPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PyramidPhase)
}

func (_m *MockPyramidGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPyramidGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPyramidGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockPyramidGame) GetPyramid() [domain.PyramidRowCnt][]*domain.PyramidCard {
	ret := _m.Called()
	return ret.Get(0).([domain.PyramidRowCnt][]*domain.PyramidCard)
}

func (_m *MockPyramidGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockPyramidGame) IsExposed(row, col int) bool {
	ret := _m.Called(row, col)
	return ret.Bool(0)
}

func (_m *MockPyramidGame) IsRemovableKing(row, col int) bool {
	args := _m.Called(row, col)
	return args.Bool(0)
}

func (_m *MockPyramidGame) IsWasteKingRemovable() bool {
	args := _m.Called()
	return args.Bool(0)
}

func (_m *MockPyramidGame) AllRemoved() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPyramidGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockPyramidGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
