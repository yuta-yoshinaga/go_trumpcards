//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCrazyQuiltGame クレイジーキルト ゲームモック
type MockCrazyQuiltGame struct {
	mock.Mock
}

func (_m *MockCrazyQuiltGame) Reset() {
	_m.Called()
}

func (_m *MockCrazyQuiltGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) MoveQuiltToFoundation(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) MoveQuiltToWaste(idx int) error {
	ret := _m.Called(idx)
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) IsAvailable(idx int) bool {
	ret := _m.Called(idx)
	return ret.Bool(0)
}

func (_m *MockCrazyQuiltGame) GetRedealsLeft() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrazyQuiltGame) IsAscendingFoundation(fIdx int) bool {
	ret := _m.Called(fIdx)
	return ret.Bool(0)
}

func (_m *MockCrazyQuiltGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) GiveUp() {
	_m.Called()
}

func (_m *MockCrazyQuiltGame) GetHint() *domain.CrazyQuiltHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CrazyQuiltHint)
}

func (_m *MockCrazyQuiltGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCrazyQuiltGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrazyQuiltGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCrazyQuiltGame) GetPhase() domain.CrazyQuiltPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CrazyQuiltPhase)
}

func (_m *MockCrazyQuiltGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrazyQuiltGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCrazyQuiltGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockCrazyQuiltGame) GetQuilt() [domain.CrazyQuiltCells]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CrazyQuiltCells]*domain.Card)
}

func (_m *MockCrazyQuiltGame) GetFoundation() [domain.CrazyQuiltFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CrazyQuiltFoundationCnt][]*domain.Card)
}

func (_m *MockCrazyQuiltGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCrazyQuiltGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockCrazyQuiltGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCrazyQuiltGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
