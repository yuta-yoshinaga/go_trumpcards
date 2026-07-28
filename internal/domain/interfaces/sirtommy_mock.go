//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSirTommyGame サー・トミーゲームモック
type MockSirTommyGame struct {
	mock.Mock
}

func (_m *MockSirTommyGame) Reset() { _m.Called() }

func (_m *MockSirTommyGame) PlayStockToFoundation(fIdx int) error {
	return _m.Called(fIdx).Error(0)
}

func (_m *MockSirTommyGame) PlayStockToWaste(wasteIdx int) error {
	return _m.Called(wasteIdx).Error(0)
}

func (_m *MockSirTommyGame) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	return _m.Called(wasteIdx, fIdx).Error(0)
}

func (_m *MockSirTommyGame) GiveUp() { _m.Called() }

func (_m *MockSirTommyGame) Undo() error { return _m.Called().Error(0) }

func (_m *MockSirTommyGame) CanUndo() bool { return _m.Called().Bool(0) }

func (_m *MockSirTommyGame) UndoN(n int) error { return _m.Called(n).Error(0) }

func (_m *MockSirTommyGame) UndoToEscape() int { return _m.Called().Int(0) }

func (_m *MockSirTommyGame) AutoComplete() error { return _m.Called().Error(0) }

func (_m *MockSirTommyGame) GetHint() *domain.SirTommyHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SirTommyHint)
}

func (_m *MockSirTommyGame) GetPhase() domain.SirTommyPhase {
	return _m.Called().Get(0).(domain.SirTommyPhase)
}

func (_m *MockSirTommyGame) GetMoveCount() int { return _m.Called().Int(0) }

func (_m *MockSirTommyGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockSirTommyGame) GetStockTop() *domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.Card)
}

func (_m *MockSirTommyGame) GetWastes() [domain.SirTommyWasteCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.SirTommyWasteCnt][]*domain.Card)
}

func (_m *MockSirTommyGame) GetFoundations() [domain.SirTommyFoundationCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.SirTommyFoundationCnt][]*domain.Card)
}

func (_m *MockSirTommyGame) IsStalemate() bool { return _m.Called().Bool(0) }

func (_m *MockSirTommyGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSirTommyGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
