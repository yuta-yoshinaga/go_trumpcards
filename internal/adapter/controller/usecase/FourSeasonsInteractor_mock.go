//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFourSeasonsInteractor フォーシーズンズインタラクターモック
type MockFourSeasonsInteractor struct {
	mock.Mock
}

func (_m *MockFourSeasonsInteractor) Reset() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) Draw() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) MoveWasteToTableau(col int) string {
	return _m.Called(col).String(0)
}

func (_m *MockFourSeasonsInteractor) MoveWasteToFoundation(fIdx int) string {
	return _m.Called(fIdx).String(0)
}

func (_m *MockFourSeasonsInteractor) MoveTableauToTableau(fromCol, toCol int) string {
	return _m.Called(fromCol, toCol).String(0)
}

func (_m *MockFourSeasonsInteractor) MoveTableauToFoundation(col, fIdx int) string {
	return _m.Called(col, fIdx).String(0)
}

func (_m *MockFourSeasonsInteractor) GiveUp() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) Hint() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) AutoComplete() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) ActionLog() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) Undo() string { return _m.Called().String(0) }

func (_m *MockFourSeasonsInteractor) UndoN(n int) string { return _m.Called(n).String(0) }

func (_m *MockFourSeasonsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
