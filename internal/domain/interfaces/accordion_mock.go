//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAccordionGame アコーディオンゲームモック
type MockAccordionGame struct {
	mock.Mock
}

func (_m *MockAccordionGame) Reset() {
	_m.Called()
}

func (_m *MockAccordionGame) Move(fromIdx, toIdx int) error {
	ret := _m.Called(fromIdx, toIdx)
	return ret.Error(0)
}

func (_m *MockAccordionGame) GiveUp() {
	_m.Called()
}

func (_m *MockAccordionGame) AutoComplete() error {
	return _m.Called().Error(0)
}

func (_m *MockAccordionGame) GetHint() *domain.AccordionHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.AccordionHint)
}

func (_m *MockAccordionGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAccordionGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAccordionGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAccordionGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockAccordionGame) GetPhase() domain.AccordionPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AccordionPhase)
}

func (_m *MockAccordionGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAccordionGame) GetPiles() [][]*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([][]*domain.Card)
}

func (_m *MockAccordionGame) GetPileCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAccordionGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAccordionGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAccordionGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
