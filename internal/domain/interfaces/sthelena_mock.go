//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockStHelenaGame セント・ヘレナ・ソリティアのモック。
type MockStHelenaGame struct {
	mock.Mock
}

func (_m *MockStHelenaGame) Reset() {
	_m.Called()
}

func (_m *MockStHelenaGame) MoveTableauToTableau(fromCol, toCol int) error {
	ret := _m.Called(fromCol, toCol)
	return ret.Error(0)
}

func (_m *MockStHelenaGame) MoveTableauToFoundation(fromCol, foundationIdx int) error {
	ret := _m.Called(fromCol, foundationIdx)
	return ret.Error(0)
}

func (_m *MockStHelenaGame) Redeal() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStHelenaGame) GiveUp() {
	_m.Called()
}

func (_m *MockStHelenaGame) GetHint() *domain.StHelenaHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.StHelenaHint)
}

func (_m *MockStHelenaGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStHelenaGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockStHelenaGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockStHelenaGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStHelenaGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockStHelenaGame) GetPhase() domain.StHelenaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.StHelenaPhase)
}

func (_m *MockStHelenaGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStHelenaGame) RestrictionsActive() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockStHelenaGame) GetRedealsRemaining() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockStHelenaGame) GetTableau() [domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.StHelenaTableauCnt][]*domain.StHelenaTableauCard)
}

func (_m *MockStHelenaGame) GetFoundation() [domain.StHelenaFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.StHelenaFoundationCnt][]*domain.Card)
}

func (_m *MockStHelenaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockStHelenaGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockStHelenaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
