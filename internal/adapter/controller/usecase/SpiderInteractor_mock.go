//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSpiderInteractor スパイダーソリティアインタラクターモック
type MockSpiderInteractor struct {
	mock.Mock
}

func (_m *MockSpiderInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) ResetWithConfig(cfg domain.SpiderConfig) string {
	ret := _m.Called(cfg)
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) Deal() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) MoveTableauToTableau(fromCol, cardIndex, toCol int) string {
	ret := _m.Called(fromCol, cardIndex, toCol)
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) GiveUp() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) Hint() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) AutoComplete() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) ActionLog() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) Undo() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

func (_m *MockSpiderInteractor) UndoN(n int) string {
	ret := _m.Called(n)
	return ret.Get(0).(string)
}
