//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockAccordionInteractor アコーディオンインタラクターモック
type MockAccordionInteractor struct {
	mock.Mock
}

func (_m *MockAccordionInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) Move(fromIdx, toIdx int) string {
	ret := _m.Called(fromIdx, toIdx)
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) AutoComplete() string {
	return _m.Called().String(0)
}

func (_m *MockAccordionInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockAccordionInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
