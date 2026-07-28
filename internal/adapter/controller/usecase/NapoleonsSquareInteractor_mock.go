//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockNapoleonsSquareInteractor ナポレオンズ・スクエア インタラクターモック
type MockNapoleonsSquareInteractor struct {
	mock.Mock
}

func (_m *MockNapoleonsSquareInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) MoveWasteToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockNapoleonsSquareInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockNapoleonsSquareInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
