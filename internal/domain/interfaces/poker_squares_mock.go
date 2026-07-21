//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPokerSquaresGame はポーカー・スクエアズゲームのモック。
type MockPokerSquaresGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockPokerSquaresGame) Reset() { _m.Called() }

// Place モック
func (_m *MockPokerSquaresGame) Place(row, col int) error {
	ret := _m.Called(row, col)
	return ret.Error(0)
}

// Undo モック
func (_m *MockPokerSquaresGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo モック
func (_m *MockPokerSquaresGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GiveUp モック
func (_m *MockPokerSquaresGame) GiveUp() { _m.Called() }

// GetPhase モック
func (_m *MockPokerSquaresGame) GetPhase() domain.PokerSquaresPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.PokerSquaresPhase)
}

// GetBoard モック
func (_m *MockPokerSquaresGame) GetBoard() [domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.PokerSquaresGridSize][domain.PokerSquaresGridSize]*domain.Card)
}

// GetCurrentCard モック
func (_m *MockPokerSquaresGame) GetCurrentCard() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

// GetPlacedCount モック
func (_m *MockPokerSquaresGame) GetPlacedCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsComplete モック
func (_m *MockPokerSquaresGame) IsComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// EvaluateRow モック
func (_m *MockPokerSquaresGame) EvaluateRow(r int) int {
	ret := _m.Called(r)
	return ret.Int(0)
}

// EvaluateCol モック
func (_m *MockPokerSquaresGame) EvaluateCol(c int) int {
	ret := _m.Called(c)
	return ret.Int(0)
}

// RowScore モック
func (_m *MockPokerSquaresGame) RowScore(r int) int {
	ret := _m.Called(r)
	return ret.Int(0)
}

// ColScore モック
func (_m *MockPokerSquaresGame) ColScore(c int) int {
	ret := _m.Called(c)
	return ret.Int(0)
}

// TotalScore モック
func (_m *MockPokerSquaresGame) TotalScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockPokerSquaresGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockPokerSquaresGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHint モック
func (_m *MockPokerSquaresGame) GetHint() *domain.PokerSquaresHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.PokerSquaresHint)
}
