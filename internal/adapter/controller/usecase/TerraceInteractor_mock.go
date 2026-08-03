//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockTerraceInteractor テラス インタラクターモック
type MockTerraceInteractor struct {
	mock.Mock
}

func (_m *MockTerraceInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) MoveReserveToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) MoveWasteToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	ret := _m.Called(fromPile, toPile)
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockTerraceInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockTerraceInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
