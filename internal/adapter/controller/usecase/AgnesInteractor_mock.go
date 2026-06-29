//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockAgnesInteractor アグネス・ソレルインタラクターモック
type MockAgnesInteractor struct {
	mock.Mock
}

func (_m *MockAgnesInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) DealStock() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockAgnesInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockAgnesInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
