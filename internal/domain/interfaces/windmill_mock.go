//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWindmillGame ウィンドミル ゲームモック
type MockWindmillGame struct {
	mock.Mock
}

func (_m *MockWindmillGame) Reset() {
	_m.Called()
}

func (_m *MockWindmillGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWindmillGame) MoveSailToCenter(sailIdx int) error {
	ret := _m.Called(sailIdx)
	return ret.Error(0)
}

func (_m *MockWindmillGame) MoveSailToCorner(sailIdx, cornerIdx int) error {
	ret := _m.Called(sailIdx, cornerIdx)
	return ret.Error(0)
}

func (_m *MockWindmillGame) MoveWasteToCenter() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWindmillGame) MoveWasteToCorner(cornerIdx int) error {
	ret := _m.Called(cornerIdx)
	return ret.Error(0)
}

func (_m *MockWindmillGame) MoveCornerToCenter(cornerIdx int) error {
	ret := _m.Called(cornerIdx)
	return ret.Error(0)
}

func (_m *MockWindmillGame) GiveUp() {
	_m.Called()
}

func (_m *MockWindmillGame) GetHint() *domain.WindmillHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.WindmillHint)
}

func (_m *MockWindmillGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWindmillGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWindmillGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWindmillGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWindmillGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockWindmillGame) GetPhase() domain.WindmillPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.WindmillPhase)
}

func (_m *MockWindmillGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWindmillGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWindmillGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockWindmillGame) GetSails() [domain.WindmillSailCnt]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.WindmillSailCnt]*domain.Card)
}

func (_m *MockWindmillGame) GetCenter() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockWindmillGame) GetCorners() [domain.WindmillCornerCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.WindmillCornerCnt][]*domain.Card)
}

func (_m *MockWindmillGame) IsTransferBlocked() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWindmillGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWindmillGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockWindmillGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockWindmillGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
