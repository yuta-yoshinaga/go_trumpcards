//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBristolGame ブリストルゲームモック
type MockBristolGame struct {
	mock.Mock
}

func (_m *MockBristolGame) Reset() { _m.Called() }

func (_m *MockBristolGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBristolGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockBristolGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockBristolGame) MoveFanToTableau(fanIdx, toCol int) error {
	ret := _m.Called(fanIdx, toCol)
	return ret.Error(0)
}

func (_m *MockBristolGame) MoveFanToFoundation(fanIdx int) error {
	ret := _m.Called(fanIdx)
	return ret.Error(0)
}

func (_m *MockBristolGame) GiveUp() { _m.Called() }

func (_m *MockBristolGame) GetHint() *domain.BristolHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.BristolHint)
}

func (_m *MockBristolGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBristolGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockBristolGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockBristolGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockBristolGame) GetPhase() domain.BristolPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.BristolPhase)
}

func (_m *MockBristolGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBristolGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockBristolGame) GetTableau() [domain.BristolTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BristolTableauCnt][]*domain.Card)
}

func (_m *MockBristolGame) GetFan() [domain.BristolFanCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BristolFanCnt][]*domain.Card)
}

func (_m *MockBristolGame) LegalTargets(fromZone string, fromCol int) ([]int, []int) {
	ret := _m.Called(fromZone, fromCol)
	tab, _ := ret.Get(0).([]int)
	found, _ := ret.Get(1).([]int)
	return tab, found
}

func (_m *MockBristolGame) GetFoundation() [domain.BristolFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.BristolFoundationCnt][]*domain.Card)
}

func (_m *MockBristolGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockBristolGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// IsStalemate モック
func (_m *MockBristolGame) IsStalemate() bool { return _m.Called().Bool(0) }

// UndoToEscape モック
func (_m *MockBristolGame) UndoToEscape() int { return _m.Called().Int(0) }
