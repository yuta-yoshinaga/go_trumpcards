//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockStreetsAndAlleysGame Streets and Alleys ゲームモック
type MockStreetsAndAlleysGame struct {
	mock.Mock
}

func (_m *MockStreetsAndAlleysGame) Reset() {
	_m.Called()
}

func (_m *MockStreetsAndAlleysGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockStreetsAndAlleysGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockStreetsAndAlleysGame) GiveUp() {
	_m.Called()
}

func (_m *MockStreetsAndAlleysGame) GetHint() *domain.StreetsAndAlleysHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.StreetsAndAlleysHint)
}

func (_m *MockStreetsAndAlleysGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStreetsAndAlleysGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStreetsAndAlleysGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockStreetsAndAlleysGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStreetsAndAlleysGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockStreetsAndAlleysGame) GetPhase() domain.StreetsAndAlleysPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.StreetsAndAlleysPhase)
}

func (_m *MockStreetsAndAlleysGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStreetsAndAlleysGame) GetTableau() [domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.StreetsAndAlleysTableauCnt][]*domain.StreetsAndAlleysTableauCard)
}

func (_m *MockStreetsAndAlleysGame) GetFoundation() [domain.StreetsAndAlleysFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.StreetsAndAlleysFoundationCnt][]*domain.Card)
}

func (_m *MockStreetsAndAlleysGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockStreetsAndAlleysGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockStreetsAndAlleysGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockStreetsAndAlleysGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
