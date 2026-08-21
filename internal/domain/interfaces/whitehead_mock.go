//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockWhiteheadGame ホワイトヘッドゲームモック
type MockWhiteheadGame struct {
	mock.Mock
}

func (_m *MockWhiteheadGame) Reset() {
	_m.Called()
}

func (_m *MockWhiteheadGame) ResetWithConfig(cfg domain.WhiteheadConfig) {
	_m.Called(cfg)
}

func (_m *MockWhiteheadGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) GiveUp() {
	_m.Called()
}

func (_m *MockWhiteheadGame) GetHint() *domain.WhiteheadHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.WhiteheadHint)
}

func (_m *MockWhiteheadGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWhiteheadGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWhiteheadGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockWhiteheadGame) GetPhase() domain.WhiteheadPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.WhiteheadPhase)
}

func (_m *MockWhiteheadGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWhiteheadGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWhiteheadGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockWhiteheadGame) GetTableau() [domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.WhiteheadTableauCnt][]*domain.WhiteheadTableauCard)
}

func (_m *MockWhiteheadGame) GetFoundation() [domain.WhiteheadFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.WhiteheadFoundationCnt][]*domain.Card)
}

func (_m *MockWhiteheadGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockWhiteheadGame) CanAutoComplete() bool {
	args := _m.Called()
	return args.Bool(0)
}

func (_m *MockWhiteheadGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

func (_m *MockWhiteheadGame) GetDrawCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWhiteheadGame) GetScore() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockWhiteheadGame) GetScoringMode() domain.WhiteheadScoringMode {
	ret := _m.Called()
	return ret.Get(0).(domain.WhiteheadScoringMode)
}

func (_m *MockWhiteheadGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockWhiteheadGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
