//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSalicLawGame サリカ法典 ゲームモック
type MockSalicLawGame struct {
	mock.Mock
}

func (_m *MockSalicLawGame) Reset() {
	_m.Called()
}

func (_m *MockSalicLawGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSalicLawGame) MoveTableauToFoundation(pile int) error {
	ret := _m.Called(pile)
	return ret.Error(0)
}

func (_m *MockSalicLawGame) MoveTableauToTableau(fromPile, toPile int) error {
	ret := _m.Called(fromPile, toPile)
	return ret.Error(0)
}

func (_m *MockSalicLawGame) GiveUp() {
	_m.Called()
}

func (_m *MockSalicLawGame) GetHint() *domain.SalicLawHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.SalicLawHint)
}

func (_m *MockSalicLawGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSalicLawGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockSalicLawGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSalicLawGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSalicLawGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockSalicLawGame) GetPhase() domain.SalicLawPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.SalicLawPhase)
}

func (_m *MockSalicLawGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSalicLawGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSalicLawGame) GetQueens() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockSalicLawGame) GetOpenPiles() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockSalicLawGame) GetTableau() [domain.SalicLawTableauCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SalicLawTableauCnt][]*domain.Card)
}

func (_m *MockSalicLawGame) GetFoundation() [domain.SalicLawFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.SalicLawFoundationCnt][]*domain.Card)
}

func (_m *MockSalicLawGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockSalicLawGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockSalicLawGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockSalicLawGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
