//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFortyThievesInteractor フォーティシーブスインタラクターモック
type MockFortyThievesInteractor struct {
	mock.Mock
}

func (_m *MockFortyThievesInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) MoveWasteToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFortyThievesInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFortyThievesInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
