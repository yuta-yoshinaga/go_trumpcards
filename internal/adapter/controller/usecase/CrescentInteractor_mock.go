//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCrescentInteractor クレセント・ソリティアのインタラクタモック。
type MockCrescentInteractor struct {
	mock.Mock
}

func (_m *MockCrescentInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) MoveTableauToFoundation(fromCol, foundationIdx int) string {
	ret := _m.Called(fromCol, foundationIdx)
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockCrescentInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	var b []byte
	if v := ret.Get(0); v != nil {
		b = v.([]byte)
	}
	return b, ret.Error(1)
}
