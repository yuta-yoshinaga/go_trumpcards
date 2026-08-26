//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockPerseveranceInteractor パーシビアランスインタラクターモック
type MockPerseveranceInteractor struct {
	mock.Mock
}

func (_m *MockPerseveranceInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) Redeal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Targets は列 col の一番下の札を置ける先を一覧する。
func (_m *MockPerseveranceInteractor) Targets(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockPerseveranceInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockPerseveranceInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
