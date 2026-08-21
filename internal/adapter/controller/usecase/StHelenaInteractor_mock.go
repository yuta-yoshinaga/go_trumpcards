//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockStHelenaInteractor セント・ヘレナ・ソリティアのインタラクタモック。
type MockStHelenaInteractor struct {
	mock.Mock
}

func (_m *MockStHelenaInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) MoveTableauToFoundation(fromCol, foundationIdx int) string {
	ret := _m.Called(fromCol, foundationIdx)
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

func (_m *MockStHelenaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	var b []byte
	if v := ret.Get(0); v != nil {
		b = v.([]byte)
	}
	return b, ret.Error(1)
}
