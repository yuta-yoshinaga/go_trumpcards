//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockLiteratureInteractor モック
type MockLiteratureInteractor struct {
	mock.Mock
}

func (_m *MockLiteratureInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockLiteratureInteractor) ResetWithConfig(cfg domain.LiteratureConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockLiteratureInteractor) Ask(to, suit, value int) string {
	return _m.Called(to, suit, value).String(0)
}

func (_m *MockLiteratureInteractor) Claim(half int, holders []int) string {
	return _m.Called(half, holders).String(0)
}

func (_m *MockLiteratureInteractor) GetConfig() domain.LiteratureConfig {
	return _m.Called().Get(0).(domain.LiteratureConfig)
}

func (_m *MockLiteratureInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockLiteratureInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
