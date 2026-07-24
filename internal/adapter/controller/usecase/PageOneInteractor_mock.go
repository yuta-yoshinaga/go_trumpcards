//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPageOneInteractor モック
type MockPageOneInteractor struct {
	mock.Mock
}

func (_m *MockPageOneInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockPageOneInteractor) ResetWithConfig(cfg domain.PageOneConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockPageOneInteractor) Play(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockPageOneInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockPageOneInteractor) Declare() string {
	return _m.Called().String(0)
}

func (_m *MockPageOneInteractor) SkipDeclare() string {
	return _m.Called().String(0)
}

func (_m *MockPageOneInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockPageOneInteractor) GetConfig() domain.PageOneConfig {
	return _m.Called().Get(0).(domain.PageOneConfig)
}

func (_m *MockPageOneInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockPageOneInteractor) Hint() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockPageOneInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
