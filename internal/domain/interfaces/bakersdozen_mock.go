//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBakersDozenGame ベーカーズダズンゲームモック
type MockBakersDozenGame struct {
	mock.Mock
}

func (_m *MockBakersDozenGame) Reset() {
	_m.Called()
}

func (_m *MockBakersDozenGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockBakersDozenGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockBakersDozenGame) GiveUp() {
	_m.Called()
}

func (_m *MockBakersDozenGame) GetHint() *domain.BakersDozenHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BakersDozenHint)
}

func (_m *MockBakersDozenGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBakersDozenGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBakersDozenGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBakersDozenGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBakersDozenGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBakersDozenGame) GetPhase() domain.BakersDozenPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BakersDozenPhase)
}

func (_m *MockBakersDozenGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBakersDozenGame) GetTableau() [domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.BakersDozenTableauCnt][]*domain.BakersDozenTableauCard)
}

// LegalTargets は列 fromCol の一番下の札を置ける先を返す。
func (_m *MockBakersDozenGame) LegalTargets(fromCol int) ([]int, []int) {
	ret := _m.Called(fromCol)
	tab, _ := ret.Get(0).([]int)
	found, _ := ret.Get(1).([]int)
	return tab, found
}

func (_m *MockBakersDozenGame) GetFoundation() [domain.BakersDozenFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BakersDozenFoundationCnt][]*domain.Card)
}

func (_m *MockBakersDozenGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBakersDozenGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockBakersDozenGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBakersDozenGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
