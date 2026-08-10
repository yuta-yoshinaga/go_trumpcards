//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockDiplomatInteractor ディプロマット インタラクターモック
type MockDiplomatInteractor struct {
	mock.Mock
}

func (_m *MockDiplomatInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	ret := _m.Called(fromPile, toPile)
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) MoveWasteToTableau(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockDiplomatInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockDiplomatInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
