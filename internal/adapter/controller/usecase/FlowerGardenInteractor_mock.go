//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFlowerGardenInteractor Flower Garden インタラクターモック
type MockFlowerGardenInteractor struct {
	mock.Mock
}

func (_m *MockFlowerGardenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) MoveReserveToTableau(reserveIdx, toCol int) string {
	ret := _m.Called(reserveIdx, toCol)
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) MoveReserveToFoundation(reserveIdx int) string {
	ret := _m.Called(reserveIdx)
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockFlowerGardenInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockFlowerGardenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
