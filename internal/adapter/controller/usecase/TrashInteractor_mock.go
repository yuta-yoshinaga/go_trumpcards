//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockTrashInteractor トラッシュインタラクターモック
type MockTrashInteractor struct {
	mock.Mock
}

func (_m *MockTrashInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) PlaceWild(pos int) string {
	ret := _m.Called(pos)
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) CpuStep() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTrashInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
