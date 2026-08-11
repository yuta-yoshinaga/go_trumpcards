//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockRoyalCotillionInteractor ロイヤルコティヨン インタラクターモック
type MockRoyalCotillionInteractor struct {
	mock.Mock
}

func (_m *MockRoyalCotillionInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) MoveReserveToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) MoveWasteToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) MoveStockToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockRoyalCotillionInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockRoyalCotillionInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
