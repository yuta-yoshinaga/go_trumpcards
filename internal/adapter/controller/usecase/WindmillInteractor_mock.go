//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockWindmillInteractor ウィンドミル インタラクターモック
type MockWindmillInteractor struct {
	mock.Mock
}

func (_m *MockWindmillInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) MoveSailToCenter(sailIdx int) string {
	ret := _m.Called(sailIdx)
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) MoveSailToCorner(sailIdx, cornerIdx int) string {
	ret := _m.Called(sailIdx, cornerIdx)
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) MoveWasteToCenter() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) MoveWasteToCorner(cornerIdx int) string {
	ret := _m.Called(cornerIdx)
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) MoveCornerToCenter(cornerIdx int) string {
	ret := _m.Called(cornerIdx)
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockWindmillInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockWindmillInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
