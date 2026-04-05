//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpiderGame スパイダーソリティアゲームモック
type MockSpiderGame struct {
	mock.Mock
}

func (_m *MockSpiderGame) Reset() {
	_m.Called()
}

func (_m *MockSpiderGame) ResetWithConfig(cfg domain.SpiderConfig) {
	_m.Called(cfg)
}

func (_m *MockSpiderGame) Deal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSpiderGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockSpiderGame) GiveUp() {
	_m.Called()
}

func (_m *MockSpiderGame) GetHint() *domain.SpiderHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SpiderHint)
}

func (_m *MockSpiderGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSpiderGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSpiderGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSpiderGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpiderGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSpiderGame) GetPhase() domain.SpiderPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SpiderPhase)
}

func (_m *MockSpiderGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpiderGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpiderGame) GetTableau() [domain.SpiderTableauCnt][]*domain.SpiderTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.SpiderTableauCnt][]*domain.SpiderTableauCard)
}

func (_m *MockSpiderGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpiderGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSpiderGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSpiderGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSpiderGame) GetDifficulty() domain.SpiderDifficulty {
	ret := _m.Called()
	return ret.Get(0).(domain.SpiderDifficulty)
}

func (_m *MockSpiderGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
