//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockGapsGame はGapsGameのテスト用モック。
type MockGapsGame struct {
	mock.Mock
}

func (_m *MockGapsGame) Reset() {
	_m.Called()
}

func (_m *MockGapsGame) Move(fromRow, fromCol, toRow, toCol int) error {
	ret := _m.Called(fromRow, fromCol, toRow, toCol)
	return ret.Error(0)
}

func (_m *MockGapsGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGapsGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockGapsGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockGapsGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGapsGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGapsGame) GiveUp() {
	_m.Called()
}

func (_m *MockGapsGame) GetHint() *domain.GapsHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.GapsHint)
}

func (_m *MockGapsGame) GetPhase() domain.GapsPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.GapsPhase)
}

func (_m *MockGapsGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGapsGame) GetGrid() [domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell {
	ret := _m.Called()
	return ret.Get(0).([domain.GapsRowCnt][domain.GapsColCnt]domain.GapsCell)
}

func (_m *MockGapsGame) GetGapNeed(row, col int) *domain.GapsGapNeed {
	args := _m.Called(row, col)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.GapsGapNeed)
}

// GetLockedPrefixLengths モック
func (_m *MockGapsGame) GetLockedPrefixLengths() [domain.GapsRowCnt]int {
	ret := _m.Called()
	return ret.Get(0).([domain.GapsRowCnt]int)
}

func (_m *MockGapsGame) GetRedealsUsed() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGapsGame) GetRedealsRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockGapsGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGapsGame) AllWon() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGapsGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockGapsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
