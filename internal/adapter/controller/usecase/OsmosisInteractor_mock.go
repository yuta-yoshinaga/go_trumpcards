//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockOsmosisInteractor オズモシスインタラクターモック
type MockOsmosisInteractor struct {
	mock.Mock
}

func (_m *MockOsmosisInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) MoveWasteToFoundation(fIdx int) string {
	ret := _m.Called(fIdx)
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) MoveReserveToFoundation(rIdx, fIdx int) string {
	ret := _m.Called(rIdx, fIdx)
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockOsmosisInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockOsmosisInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
