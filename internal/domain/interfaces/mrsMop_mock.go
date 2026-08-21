//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMrsMopGame ミセス・モップソリティアゲームモック
type MockMrsMopGame struct {
	mock.Mock
}

func (_m *MockMrsMopGame) Reset() {
	_m.Called()
}

func (_m *MockMrsMopGame) ResetWithConfig(cfg domain.MrsMopConfig) {
	_m.Called(cfg)
}

func (_m *MockMrsMopGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockMrsMopGame) GiveUp() {
	_m.Called()
}

func (_m *MockMrsMopGame) GetHint() *domain.MrsMopHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.MrsMopHint)
}

func (_m *MockMrsMopGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMrsMopGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockMrsMopGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMrsMopGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMrsMopGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockMrsMopGame) GetPhase() domain.MrsMopPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.MrsMopPhase)
}

func (_m *MockMrsMopGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMrsMopGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMrsMopGame) GetTableau() [domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.MrsMopTableauCnt][]*domain.MrsMopTableauCard)
}

func (_m *MockMrsMopGame) GetCompletedSuits() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMrsMopGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockMrsMopGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockMrsMopGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockMrsMopGame) GetDifficulty() domain.MrsMopDifficulty {
	ret := _m.Called()
	return ret.Get(0).(domain.MrsMopDifficulty)
}

func (_m *MockMrsMopGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockMrsMopGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
