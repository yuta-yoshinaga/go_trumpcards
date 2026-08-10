//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockFourSeasonsGame フォーシーズンズゲームモック
type MockFourSeasonsGame struct {
	mock.Mock
}

func (_m *MockFourSeasonsGame) Reset() { _m.Called() }

func (_m *MockFourSeasonsGame) Draw() error { return _m.Called().Error(0) }

func (_m *MockFourSeasonsGame) MoveWasteToTableau(col int) error {
	return _m.Called(col).Error(0)
}

func (_m *MockFourSeasonsGame) MoveWasteToFoundation(fIdx int) error {
	return _m.Called(fIdx).Error(0)
}

func (_m *MockFourSeasonsGame) MoveTableauToTableau(fromCol, toCol int) error {
	return _m.Called(fromCol, toCol).Error(0)
}

func (_m *MockFourSeasonsGame) MoveTableauToFoundation(col, fIdx int) error {
	return _m.Called(col, fIdx).Error(0)
}

func (_m *MockFourSeasonsGame) GiveUp() { _m.Called() }

func (_m *MockFourSeasonsGame) GetHint() *domain.FourSeasonsHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.FourSeasonsHint)
}

func (_m *MockFourSeasonsGame) AutoComplete() error { return _m.Called().Error(0) }

func (_m *MockFourSeasonsGame) Undo() error { return _m.Called().Error(0) }

func (_m *MockFourSeasonsGame) CanUndo() bool { return _m.Called().Bool(0) }

func (_m *MockFourSeasonsGame) UndoN(n int) error { return _m.Called(n).Error(0) }

func (_m *MockFourSeasonsGame) GetPhase() domain.FourSeasonsPhase {
	return _m.Called().Get(0).(domain.FourSeasonsPhase)
}

func (_m *MockFourSeasonsGame) GetMoveCount() int { return _m.Called().Int(0) }

func (_m *MockFourSeasonsGame) GetStockCount() int { return _m.Called().Int(0) }

func (_m *MockFourSeasonsGame) GetBaseRank() int { return _m.Called().Int(0) }

func (_m *MockFourSeasonsGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockFourSeasonsGame) GetTableau() [domain.FourSeasonsTableauCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.FourSeasonsTableauCnt][]*domain.Card)
}

func (_m *MockFourSeasonsGame) GetFoundations() [domain.FourSeasonsFoundationCnt][]*domain.Card {
	return _m.Called().Get(0).([domain.FourSeasonsFoundationCnt][]*domain.Card)
}

func (_m *MockFourSeasonsGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockFourSeasonsGame) GetGameEndFlag() bool { return _m.Called().Bool(0) }
