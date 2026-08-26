//go:build test

package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockAlaskaGame アラスカゲームモック
type MockAlaskaGame struct {
	mock.Mock
}

func (_m *MockAlaskaGame) Reset() {
	_m.Called()
}

func (_m *MockAlaskaGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockAlaskaGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockAlaskaGame) GiveUp() {
	_m.Called()
}

func (_m *MockAlaskaGame) GetHint() *domain.AlaskaHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.AlaskaHint)
}

func (_m *MockAlaskaGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAlaskaGame) Undo() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockAlaskaGame) CanUndo() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAlaskaGame) UndoToEscape() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAlaskaGame) UndoN(n int) error {
	ret := _m.Called(n)
	return ret.Error(0)
}

func (_m *MockAlaskaGame) GetPhase() domain.AlaskaPhase {
	ret := _m.Called()
	return ret.Get(0).(domain.AlaskaPhase)
}

func (_m *MockAlaskaGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockAlaskaGame) GetTableau() [domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.AlaskaTableauCnt][]*domain.AlaskaTableauCard)
}

func (_m *MockAlaskaGame) GetFoundation() [domain.AlaskaFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.AlaskaFoundationCnt][]*domain.Card)
}

func (_m *MockAlaskaGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAlaskaGame) IsStalemate() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockAlaskaGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}

// GetGameEndFlag mocks the GetGameEndFlag call.
func (_m *MockAlaskaGame) GetGameEndFlag() bool {
	ret := _m.Called()
	return ret.Bool(0)
}
