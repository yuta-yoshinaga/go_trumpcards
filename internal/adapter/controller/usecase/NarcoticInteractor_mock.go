//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockNarcoticInteractor ナルコティックインタラクターモック
type MockNarcoticInteractor struct {
	mock.Mock
}

func (_m *MockNarcoticInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) Remove() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Redeal モック
func (_m *MockNarcoticInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) Move(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNarcoticInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockNarcoticInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
