//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCribbageSquaresGame はクリベッジ・スクエアズゲームのモック。
type MockCribbageSquaresGame struct {
	mock.Mock
}

// Reset モック
func (_m *MockCribbageSquaresGame) Reset() { _m.Called() }

// Place モック
func (_m *MockCribbageSquaresGame) Place(row, col int) error {
	ret := _m.Called(row, col)
	return ret.Error(0)
}

// Undo モック
func (_m *MockCribbageSquaresGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

// CanUndo モック
func (_m *MockCribbageSquaresGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GiveUp モック
func (_m *MockCribbageSquaresGame) GiveUp() { _m.Called() }

// GetPhase モック
func (_m *MockCribbageSquaresGame) GetPhase() domain.CribbageSquaresPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CribbageSquaresPhase)
}

// GetBoard モック
func (_m *MockCribbageSquaresGame) GetBoard() [domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CribbageSquaresGridSize][domain.CribbageSquaresGridSize]*domain.Card)
}

// GetCurrentCard モック
func (_m *MockCribbageSquaresGame) GetCurrentCard() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

// GetPlacedCount モック
func (_m *MockCribbageSquaresGame) GetPlacedCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

// IsComplete モック
func (_m *MockCribbageSquaresGame) IsComplete() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// RowScore モック
func (_m *MockCribbageSquaresGame) RowScore(r int) int {
	ret := _m.Called(r)
	return ret.Int(0)
}

// ColScore モック
func (_m *MockCribbageSquaresGame) ColScore(c int) int {
	ret := _m.Called(c)
	return ret.Int(0)
}

// GetStarter モック
func (_m *MockCribbageSquaresGame) GetStarter() *domain.Card {
	ret := _m.Called()
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(*domain.Card)
}

// RowPartialDetail モック
func (_m *MockCribbageSquaresGame) RowPartialDetail(r int) domain.CribbageScoreDetail {
	return _m.Called(r).Get(0).(domain.CribbageScoreDetail)
}

// ColPartialDetail モック
func (_m *MockCribbageSquaresGame) ColPartialDetail(col int) domain.CribbageScoreDetail {
	return _m.Called(col).Get(0).(domain.CribbageScoreDetail)
}

// RowDetail モック
func (_m *MockCribbageSquaresGame) RowDetail(r int) domain.CribbageScoreDetail {
	ret := _m.Called(r)
	return ret.Get(0).(domain.CribbageScoreDetail)
}

// ColDetail モック
func (_m *MockCribbageSquaresGame) ColDetail(c int) domain.CribbageScoreDetail {
	ret := _m.Called(c)
	return ret.Get(0).(domain.CribbageScoreDetail)
}

// IsWin モック
func (_m *MockCribbageSquaresGame) IsWin() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// TotalScore モック
func (_m *MockCribbageSquaresGame) TotalScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

// GetActionLog モック
func (_m *MockCribbageSquaresGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCribbageSquaresGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetHint モック
func (_m *MockCribbageSquaresGame) GetHint() *domain.CribbageSquaresHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CribbageSquaresHint)
}
