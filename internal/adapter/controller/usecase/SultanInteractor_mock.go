//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockSultanInteractor スルタンインタラクターモック
type MockSultanInteractor struct {
	mock.Mock
}

func (_m *MockSultanInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) MoveDivanToFoundation(divanIdx int) string {
	ret := _m.Called(divanIdx)
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSultanInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSultanInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
