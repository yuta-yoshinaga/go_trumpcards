//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTrashGame トラッシュゲームモック
type MockTrashGame struct {
	mock.Mock
}

func (_m *MockTrashGame) Reset() {
	_m.Called()
}

func (_m *MockTrashGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTrashGame) PlaceWild(pos int) error {
	ret := _m.Called(pos)
	return ret.Error(0)
}

func (_m *MockTrashGame) CpuStep() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockTrashGame) IsCpuTurn() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockTrashGame) IsCpuPlayer(idx int) bool {
	ret := _m.Called(idx)
	return ret.Bool(0)
}

func (_m *MockTrashGame) GetPhase() domain.TrashPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.TrashPhase)
}

func (_m *MockTrashGame) GetCurrent() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTrashGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTrashGame) GetStockSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTrashGame) GetDiscardSize() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTrashGame) GetDiscardTop() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

func (_m *MockTrashGame) GetPending() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

func (_m *MockTrashGame) GetPlayerSlots(idx int) []domain.TrashSlot {
	ret := _m.Called(idx)
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]domain.TrashSlot)
}

func (_m *MockTrashGame) GetWinner() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockTrashGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
