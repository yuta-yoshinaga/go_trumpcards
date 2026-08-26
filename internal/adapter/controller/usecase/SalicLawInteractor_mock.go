//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockSalicLawInteractor サリカ法典 インタラクターモック
type MockSalicLawInteractor struct {
	mock.Mock
}

func (_m *MockSalicLawInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) MoveTableauToFoundation(pile int) string {
	ret := _m.Called(pile)
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) MoveTableauToTableau(fromPile, toPile int) string {
	ret := _m.Called(fromPile, toPile)
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSalicLawInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}

// Snapshot モック
func (_m *MockSalicLawInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
