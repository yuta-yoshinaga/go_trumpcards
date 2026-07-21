//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockWaspInteractor ワスプインタラクターモック
type MockWaspInteractor struct {
	mock.Mock
}

func (_m *MockWaspInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) LegalMoves(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockWaspInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
