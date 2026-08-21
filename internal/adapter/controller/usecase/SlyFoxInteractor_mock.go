//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockSlyFoxInteractor スライ・フォックス インタラクターモック
type MockSlyFoxInteractor struct {
	mock.Mock
}

func (_m *MockSlyFoxInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) DealToPile(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) DealToFoundation(fIdx int) string {
	ret := _m.Called(fIdx)
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSlyFoxInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSlyFoxInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
