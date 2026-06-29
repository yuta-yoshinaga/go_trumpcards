//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKingAlbertGame King Albert ゲームモック
type MockKingAlbertGame struct {
	mock.Mock
}

func (_m *MockKingAlbertGame) Reset() {
	_m.Called()
}

func (_m *MockKingAlbertGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) MoveReserveToTableau(reserveIdx, toCol int) error {
	ret := _m.Called(reserveIdx, toCol)
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) MoveReserveToFoundation(reserveIdx int) error {
	ret := _m.Called(reserveIdx)
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) GiveUp() {
	_m.Called()
}

func (_m *MockKingAlbertGame) GetHint() *domain.KingAlbertHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.KingAlbertHint)
}

func (_m *MockKingAlbertGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockKingAlbertGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockKingAlbertGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockKingAlbertGame) GetPhase() domain.KingAlbertPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KingAlbertPhase)
}

func (_m *MockKingAlbertGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockKingAlbertGame) GetTableau() [domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.KingAlbertTableauCnt][]*domain.KingAlbertTableauCard)
}

func (_m *MockKingAlbertGame) GetReserve() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockKingAlbertGame) GetFoundation() [domain.KingAlbertFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.KingAlbertFoundationCnt][]*domain.Card)
}

func (_m *MockKingAlbertGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockKingAlbertGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockKingAlbertGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockKingAlbertGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
