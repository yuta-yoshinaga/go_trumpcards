//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAuldLangSyneGame オールド・ラング・サインゲームモック
type MockAuldLangSyneGame struct {
	mock.Mock
}

func (_m *MockAuldLangSyneGame) Reset() { _m.Called() }

func (_m *MockAuldLangSyneGame) Deal() error { return _m.Called().Error(0) }

func (_m *MockAuldLangSyneGame) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	return _m.Called(wasteIdx, fIdx).Error(0)
}

func (_m *MockAuldLangSyneGame) GiveUp() { _m.Called() }

func (_m *MockAuldLangSyneGame) Undo() error { return _m.Called().Error(0) }

func (_m *MockAuldLangSyneGame) CanUndo() bool { return _m.Called().Bool(0) }

func (_m *MockAuldLangSyneGame) UndoN(n int) error { return _m.Called(n).Error(0) }

func (_m *MockAuldLangSyneGame) UndoToEscape() int { return _m.Called().Int(0) }

func (_m *MockAuldLangSyneGame) AutoComplete() error { return _m.Called().Error(0) }

func (_m *MockAuldLangSyneGame) GetHint() *domain.AuldLangSyneHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.AuldLangSyneHint)
}

func (_m *MockAuldLangSyneGame) GetPhase() domain.AuldLangSynePhase {
	return _m.Called().Get(0).(domain.AuldLangSynePhase)
}

func (_m *MockAuldLangSyneGame) GetMoveCount() int { return _m.Called().Int(0) }

func (_m *MockAuldLangSyneGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockAuldLangSyneGame) GetWastes() [domain.AuldLangSyneWasteCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.AuldLangSyneWasteCnt][]*domain.Card)
}

func (_m *MockAuldLangSyneGame) GetFoundations() [domain.AuldLangSyneFoundationCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.AuldLangSyneFoundationCnt][]*domain.Card)
}

func (_m *MockAuldLangSyneGame) IsStalemate() bool { return _m.Called().Bool(0) }

func (_m *MockAuldLangSyneGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAuldLangSyneGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
