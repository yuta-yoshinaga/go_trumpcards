//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockMissMilliganInteractor ミス・ミリガン インタラクターモック
type MockMissMilliganInteractor struct {
	mock.Mock
}

func (_m *MockMissMilliganInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) Waive(col, cardIndex int) string {
	ret := _m.Called(col, cardIndex)
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) PlaceWaived(toCol int) string {
	ret := _m.Called(toCol)
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) MoveWaivedToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMissMilliganInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMissMilliganInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
