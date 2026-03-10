package interfaces

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKlondikeGame クロンダイクゲームモック
type MockKlondikeGame struct {
	mock.Mock
}

func (_m *MockKlondikeGame) Reset() {
	_m.Called()
}

func (_m *MockKlondikeGame) Draw() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockKlondikeGame) MoveWasteToTableau(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockKlondikeGame) MoveWasteToFoundation() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockKlondikeGame) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Error(0)
}

func (_m *MockKlondikeGame) MoveTableauToFoundation(col int) error {
	ret := _m.Called(col)
	return ret.Error(0)
}

func (_m *MockKlondikeGame) GiveUp() {
	_m.Called()
}

func (_m *MockKlondikeGame) GetHint() *domain.KlondikeHint {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.(*domain.KlondikeHint)
}

func (_m *MockKlondikeGame) AutoComplete() error {
	ret := _m.Called()
	return ret.Error(0)
}

func (_m *MockKlondikeGame) GetPhase() domain.KlondikePhase {
	ret := _m.Called()
	return ret.Get(0).(domain.KlondikePhase)
}

func (_m *MockKlondikeGame) GetMoveCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockKlondikeGame) GetStockCount() int {
	ret := _m.Called()
	return ret.Int(0)
}

func (_m *MockKlondikeGame) GetWaste() []*domain.Card {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.Card)
}

func (_m *MockKlondikeGame) GetTableau() [domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard {
	ret := _m.Called()
	return ret.Get(0).([domain.KlondikeTableauCnt][]*domain.KlondikeTableauCard)
}

func (_m *MockKlondikeGame) GetFoundation() [domain.KlondikeFoundationCnt][]*domain.Card {
	ret := _m.Called()
	return ret.Get(0).([domain.KlondikeFoundationCnt][]*domain.Card)
}

func (_m *MockKlondikeGame) AllFaceUp() bool {
	ret := _m.Called()
	return ret.Bool(0)
}

func (_m *MockKlondikeGame) GetActionLog() []*domain.ActionLogEntry {
	ret := _m.Called()
	v := ret.Get(0)
	if v == nil {
		return nil
	}
	return v.([]*domain.ActionLogEntry)
}
