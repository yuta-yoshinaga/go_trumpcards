//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockMatrimonyInteractor マトリモニー インタラクターモック
type MockMatrimonyInteractor struct {
	mock.Mock
}

func (_m *MockMatrimonyInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) MoveWasteToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) MoveStockToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockMatrimonyInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockMatrimonyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
