//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockOsmosisGame オズモシスゲームモック
type MockOsmosisGame struct {
	mock.Mock
}

func (_m *MockOsmosisGame) Reset() { _m.Called() }

func (_m *MockOsmosisGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockOsmosisGame) MoveWasteToFoundation(fIdx int) error {
	ret := _m.Called(fIdx)
	return ret.Error(0)
}

func (_m *MockOsmosisGame) MoveReserveToFoundation(rIdx, fIdx int) error {
	ret := _m.Called(rIdx, fIdx)
	return ret.Error(0)
}

func (_m *MockOsmosisGame) GiveUp() { _m.Called() }

func (_m *MockOsmosisGame) GetHint() *domain.OsmosisHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.OsmosisHint)
}

func (_m *MockOsmosisGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockOsmosisGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockOsmosisGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockOsmosisGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockOsmosisGame) GetPhase() domain.OsmosisPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.OsmosisPhase)
}

func (_m *MockOsmosisGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockOsmosisGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockOsmosisGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockOsmosisGame) GetReserve() [domain.OsmosisReserveCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.OsmosisReserveCnt][]*domain.Card)
}

func (_m *MockOsmosisGame) GetFoundation() [domain.OsmosisFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.OsmosisFoundationCnt][]*domain.Card)
}

func (_m *MockOsmosisGame) GetBaseRank() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockOsmosisGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockOsmosisGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockOsmosisGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
