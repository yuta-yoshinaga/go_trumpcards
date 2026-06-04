//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockYanivInteractor モック
type MockYanivInteractor struct {
	mock.Mock
}

func (_m *MockYanivInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockYanivInteractor) ResetWithConfig(cfg domain.YanivConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockYanivInteractor) Discard(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockYanivInteractor) DeclareYaniv() string {
	return _m.Called().String(0)
}

func (_m *MockYanivInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockYanivInteractor) DrawFromPickup(end int) string {
	return _m.Called(end).String(0)
}

func (_m *MockYanivInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockYanivInteractor) GetConfig() domain.YanivConfig {
	return _m.Called().Get(0).(domain.YanivConfig)
}

func (_m *MockYanivInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockYanivInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
