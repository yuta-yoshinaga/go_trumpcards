//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSultanGame スルタンゲームモック
type MockSultanGame struct {
	mock.Mock
}

func (_m *MockSultanGame) Reset() {
	_m.Called()
}

func (_m *MockSultanGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSultanGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSultanGame) MoveDivanToFoundation(divanIdx int) error {
	ret := _m.Called(divanIdx)
	return ret.Error(0)
}

func (_m *MockSultanGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSultanGame) GiveUp() {
	_m.Called()
}

func (_m *MockSultanGame) GetHint() *domain.SultanHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SultanHint)
}

func (_m *MockSultanGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSultanGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSultanGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSultanGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSultanGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSultanGame) GetPhase() domain.SultanPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SultanPhase)
}

func (_m *MockSultanGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSultanGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSultanGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockSultanGame) GetDivan() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockSultanGame) GetFoundation() [domain.SultanFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SultanFoundationCnt][]*domain.Card)
}

func (_m *MockSultanGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSultanGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSultanGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetRedealCount mocks the GetRedealCount call.
func (_m *MockSultanGame) GetRedealCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// CanRedeal mocks the CanRedeal call.
func (_m *MockSultanGame) CanRedeal() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSultanGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
