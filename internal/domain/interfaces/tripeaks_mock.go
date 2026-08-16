//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTriPeaksGame トリピークスゲームモック
type MockTriPeaksGame struct {
	mock.Mock
}

func (_m *MockTriPeaksGame) Reset() {
	_m.Called()
}

func (_m *MockTriPeaksGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTriPeaksGame) Remove(row, col int) error {
	ret := _m.Called(row, col)
	return ret.Error(0)
}

func (_m *MockTriPeaksGame) GiveUp() {
	_m.Called()
}

func (_m *MockTriPeaksGame) GetHint() *domain.TriPeaksHint {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.(*domain.TriPeaksHint)
}

func (_m *MockTriPeaksGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTriPeaksGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockTriPeaksGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTriPeaksGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockTriPeaksGame) GetPhase() domain.TriPeaksPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TriPeaksPhase)
}

func (_m *MockTriPeaksGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockTriPeaksGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTriPeaksGame) GetCombo() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTriPeaksGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Get(0).(int)
}

func (_m *MockTriPeaksGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.Card)
}

func (_m *MockTriPeaksGame) GetLayout() [domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard {
	ret := _m.Called()
	return ret.Get(0).([domain.TriPeaksRowCnt][domain.TriPeaksColCnt]*domain.TriPeaksCard)
}

func (_m *MockTriPeaksGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	r0 := ret.Get(0)
	if r0 == nil {
		return nil
	}
	return r0.([]*domain.ActionLogEntry)
}

func (_m *MockTriPeaksGame) IsExposed(row, col int) bool {
	ret := _m.Called(row, col)
	return ret.Get(0).(bool)
}

func (_m *MockTriPeaksGame) AllRemoved() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

func (_m *MockTriPeaksGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Get(0).(bool)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockTriPeaksGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
