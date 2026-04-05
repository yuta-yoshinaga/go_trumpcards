//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKlondikeInteractor クロンダイクインタラクターモック
type MockKlondikeInteractor struct {
	mock.Mock
}

func (_m *MockKlondikeInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) ResetWithConfig(cfg domain.KlondikeConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) Draw() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) MoveWasteToTableau(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) MoveWasteToFoundation() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) MoveTableauToFoundation(col int) string {
	ret := _m.Called(col)
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockKlondikeInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}
