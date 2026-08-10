//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCalculationGame カルキュレーションゲームモック
type MockCalculationGame struct {
	mock.Mock
}

func (_m *MockCalculationGame) Reset() { _m.Called() }

func (_m *MockCalculationGame) PlayStockToFoundation(fIdx int) error {
	return _m.Called(fIdx).Error(0)
}

func (_m *MockCalculationGame) PlayStockToWaste(wasteIdx int) error {
	return _m.Called(wasteIdx).Error(0)
}

func (_m *MockCalculationGame) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	return _m.Called(wasteIdx, fIdx).Error(0)
}

func (_m *MockCalculationGame) GiveUp() { _m.Called() }

func (_m *MockCalculationGame) Undo() error { return _m.Called().Error(0) }

func (_m *MockCalculationGame) CanUndo() bool { return _m.Called().Bool(0) }

func (_m *MockCalculationGame) UndoN(n int) error { return _m.Called(n).Error(0) }

func (_m *MockCalculationGame) UndoToEscape() int { return _m.Called().Int(0) }

func (_m *MockCalculationGame) AutoComplete() error { return _m.Called().Error(0) }

func (_m *MockCalculationGame) GetHint() *domain.CalculationHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CalculationHint)
}

func (_m *MockCalculationGame) GetPhase() domain.CalculationPhase {
	return _m.Called().Get(0).(domain.CalculationPhase)
}

func (_m *MockCalculationGame) GetMoveCount() int { return _m.Called().Int(0) }

func (_m *MockCalculationGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockCalculationGame) GetStockTop() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

func (_m *MockCalculationGame) GetWastes() [domain.CalculationWasteCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.CalculationWasteCnt][]*domain.Card)
}

func (_m *MockCalculationGame) GetFoundations() [domain.CalculationFoundationCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.CalculationFoundationCnt][]*domain.Card)
}

func (_m *MockCalculationGame) GetNextFoundationRank(fIdx int) int {
	args := _m.Called(fIdx)
	return args.Int(0)
}

func (_m *MockCalculationGame) IsStalemate() bool { return _m.Called().Bool(0) }

func (_m *MockCalculationGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCalculationGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
