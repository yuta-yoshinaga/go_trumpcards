//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockBigBenInteractor ビッグ・ベン インタラクターモック
type MockBigBenInteractor struct {
	mock.Mock
}

func (_m *MockBigBenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) MoveTableauToFoundation(col, fIdx int) string {
	ret := _m.Called(col, fIdx)
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	ret := _m.Called(fromCol, toCol)
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBigBenInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBigBenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
