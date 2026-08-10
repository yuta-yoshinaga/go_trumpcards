//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockAuldLangSyneInteractor オールド・ラング・サインインタラクターモック
type MockAuldLangSyneInteractor struct {
	mock.Mock
}

func (_m *MockAuldLangSyneInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) Deal() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) PlayWasteToFoundation(wasteIdx, fIdx int) string {
	return _m.Called(wasteIdx, fIdx).String(0)
}

func (_m *MockAuldLangSyneInteractor) GiveUp() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) AutoComplete() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) Undo() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) UndoN(n int) string {
	return _m.Called(n).String(0)
}

func (_m *MockAuldLangSyneInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) ActionLog() string {
	return _m.Called().String(0)
}

func (_m *MockAuldLangSyneInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
