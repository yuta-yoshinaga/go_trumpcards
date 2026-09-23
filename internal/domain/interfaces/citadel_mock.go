//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCitadelGame Citadel ゲームモック
type MockCitadelGame struct {
	mock.Mock
}

func (_m *MockCitadelGame) Reset() {
	_m.Called()
}

func (_m *MockCitadelGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockCitadelGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockCitadelGame) GiveUp() {
	_m.Called()
}

func (_m *MockCitadelGame) GetHint() *domain.CitadelHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.CitadelHint)
}

func (_m *MockCitadelGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCitadelGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockCitadelGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCitadelGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCitadelGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockCitadelGame) GetPhase() domain.CitadelPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.CitadelPhase)
}

func (_m *MockCitadelGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockCitadelGame) GetTableau() [domain.CitadelTableauCnt][]*domain.CitadelTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.CitadelTableauCnt][]*domain.CitadelTableauCard)
}

func (_m *MockCitadelGame) GetFoundation() [domain.CitadelFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.CitadelFoundationCnt][]*domain.Card)
}

func (_m *MockCitadelGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockCitadelGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockCitadelGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockCitadelGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
