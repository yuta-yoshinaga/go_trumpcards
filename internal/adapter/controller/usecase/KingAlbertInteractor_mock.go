//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockKingAlbertInteractor King Albert インタラクターモック
type MockKingAlbertInteractor struct {
	mock.Mock
}

func (_m *MockKingAlbertInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) MoveReserveToTableau(reserveIdx, toCol int) string {
	ret := _m.Called(reserveIdx, toCol)
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) MoveReserveToFoundation(reserveIdx int) string {
	ret := _m.Called(reserveIdx)
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKingAlbertInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockKingAlbertInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
