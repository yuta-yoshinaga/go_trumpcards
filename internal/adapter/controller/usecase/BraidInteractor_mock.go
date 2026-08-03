//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockBraidInteractor ブレイド インタラクターモック
type MockBraidInteractor struct {
	mock.Mock
}

func (_m *MockBraidInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) ChooseDirection(ascending bool) string {
	ret := _m.Called(ascending)
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) MoveBraidToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) MoveFieldToFoundation(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) MoveHelperToFoundation(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) MoveWasteToHelper(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBraidInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBraidInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
