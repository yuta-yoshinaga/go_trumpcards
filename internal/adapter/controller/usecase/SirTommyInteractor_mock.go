//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockSirTommyInteractor サー・トミーインタラクターモック
type MockSirTommyInteractor struct {
	mock.Mock
}

func (_m *MockSirTommyInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) PlayStockToFoundation(fIdx int) string {
	return _m.Called(fIdx).String(0)
}

func (_m *MockSirTommyInteractor) PlayStockToWaste(wasteIdx int) string {
	return _m.Called(wasteIdx).String(0)
}

func (_m *MockSirTommyInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return _m.Called(wasteIdx, fIdx).String(0)
}

func (_m *MockSirTommyInteractor) GiveUp() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) AutoComplete() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) Undo() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) UndoN(n int) string {
	return _m.Called(n).String(0)
}

func (_m *MockSirTommyInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) ActionLog() string {
	return _m.Called().String(0)
}

func (_m *MockSirTommyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
