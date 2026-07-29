//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockDuchessInteractor ダッチェス インタラクターモック
type MockDuchessInteractor struct {
	mock.Mock
}

func (_m *MockDuchessInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) ChooseBaseRank(fanIdx int) string {
	ret := _m.Called(fanIdx)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveReserveToFoundation(fanIdx int) string {
	ret := _m.Called(fanIdx)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveReserveToTableau(fanIdx, col int) string {
	ret := _m.Called(fanIdx, col)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveWasteToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDuchessInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockDuchessInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
