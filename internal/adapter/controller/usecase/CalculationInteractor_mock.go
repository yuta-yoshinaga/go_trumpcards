//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockCalculationInteractor カルキュレーションインタラクターモック
type MockCalculationInteractor struct {
	mock.Mock
}

func (_m *MockCalculationInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) PlayStockToFoundation(fIdx int) string {
	return _m.Called(fIdx).String(0)
}

func (_m *MockCalculationInteractor) PlayStockToWaste(wasteIdx int) string {
	return _m.Called(wasteIdx).String(0)
}

func (_m *MockCalculationInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return _m.Called(wasteIdx, fIdx).String(0)
}

func (_m *MockCalculationInteractor) GiveUp() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) AutoComplete() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) Undo() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) UndoN(n int) string {
	return _m.Called(n).String(0)
}

func (_m *MockCalculationInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) ActionLog() string {
	return _m.Called().String(0)
}

func (_m *MockCalculationInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
