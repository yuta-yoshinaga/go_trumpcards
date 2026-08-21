//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRankAndFileGame ランク・アンド・ファイルゲームモック
type MockRankAndFileGame struct {
	mock.Mock
}

func (_m *MockRankAndFileGame) Reset() {
	_m.Called()
}

func (_m *MockRankAndFileGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) GiveUp() {
	_m.Called()
}

func (_m *MockRankAndFileGame) GetHint() *domain.RankAndFileHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.RankAndFileHint)
}

func (_m *MockRankAndFileGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRankAndFileGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRankAndFileGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockRankAndFileGame) GetPhase() domain.RankAndFilePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.RankAndFilePhase)
}

func (_m *MockRankAndFileGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRankAndFileGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockRankAndFileGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockRankAndFileGame) GetTableau() [domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.RankAndFileTableauCnt][]*domain.RankAndFileTableauCard)
}

func (_m *MockRankAndFileGame) GetFoundation() [domain.RankAndFileFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.RankAndFileFoundationCnt][]*domain.Card)
}

func (_m *MockRankAndFileGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockRankAndFileGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockRankAndFileGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockRankAndFileGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
