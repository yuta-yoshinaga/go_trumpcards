//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockPyramidInteractor ピラミッドインタラクターモック
type MockPyramidInteractor struct {
	mock.Mock
}

func (_m *MockPyramidInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) RemovePair(row1, col1, row2, col2 int) string {
	ret := _m.Called(row1, col1, row2, col2)
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) RemoveKing(row, col int) string {
	ret := _m.Called(row, col)
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) RemoveWithWaste(row, col int) string {
	ret := _m.Called(row, col)
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) RemoveWasteKing() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPyramidInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}
