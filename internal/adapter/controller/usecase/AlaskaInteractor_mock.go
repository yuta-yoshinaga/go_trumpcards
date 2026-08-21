//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockAlaskaInteractor アラスカインタラクターモック
type MockAlaskaInteractor struct {
	mock.Mock
}

func (_m *MockAlaskaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockAlaskaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
