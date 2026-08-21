//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPerseveranceGame パーシビアランスゲームモック
type MockPerseveranceGame struct {
	mock.Mock
}

func (_m *MockPerseveranceGame) Reset() {
	_m.Called()
}

func (_m *MockPerseveranceGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockPerseveranceGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockPerseveranceGame) GiveUp() {
	_m.Called()
}

func (_m *MockPerseveranceGame) GetHint() *domain.PerseveranceHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.PerseveranceHint)
}

func (_m *MockPerseveranceGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPerseveranceGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockPerseveranceGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPerseveranceGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPerseveranceGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockPerseveranceGame) GetPhase() domain.PerseverancePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PerseverancePhase)
}

func (_m *MockPerseveranceGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockPerseveranceGame) GetTableau() [domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.PerseveranceTableauCnt][]*domain.PerseveranceTableauCard)
}

// LegalTargets は列 fromCol の一番下の札を置ける先を返す。
func (_m *MockPerseveranceGame) LegalTargets(fromCol int) ([]int, []int) {
	ret := _m.Called(fromCol)
	tab, _ := ret.Get(0).([]int)
	found, _ := ret.Get(1).([]int)
	return tab, found
}

func (_m *MockPerseveranceGame) GetFoundation() [domain.PerseveranceFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.PerseveranceFoundationCnt][]*domain.Card)
}

func (_m *MockPerseveranceGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockPerseveranceGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockPerseveranceGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// Redeal mocks the Redeal call.
func (_m *MockPerseveranceGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

// GetRedealsLeft mocks the GetRedealsLeft call.
func (_m *MockPerseveranceGame) GetRedealsLeft() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockPerseveranceGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
