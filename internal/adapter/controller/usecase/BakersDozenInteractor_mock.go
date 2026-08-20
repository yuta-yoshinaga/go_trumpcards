//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockBakersDozenInteractor ベーカーズダズンインタラクターモック
type MockBakersDozenInteractor struct {
	mock.Mock
}

func (_m *MockBakersDozenInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Targets は列 col の一番下の札を置ける先を一覧する。
func (_m *MockBakersDozenInteractor) Targets(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockBakersDozenInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockBakersDozenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
