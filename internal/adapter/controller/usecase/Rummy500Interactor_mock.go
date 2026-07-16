//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockRummy500Interactor モック
type MockRummy500Interactor struct {
	mock.Mock
}

func (_m *MockRummy500Interactor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockRummy500Interactor) ResetWithConfig(cfg domain.Rummy500Config) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockRummy500Interactor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockRummy500Interactor) DrawFromDiscard(idx int) string {
	return _m.Called(idx).String(0)
}

func (_m *MockRummy500Interactor) Meld(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockRummy500Interactor) Layoff(meldOwner, meldIdx, cardIndex int) string {
	return _m.Called(meldOwner, meldIdx, cardIndex).String(0)
}

func (_m *MockRummy500Interactor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockRummy500Interactor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockRummy500Interactor) GetConfig() domain.Rummy500Config {
	return _m.Called().Get(0).(domain.Rummy500Config)
}

func (_m *MockRummy500Interactor) ActionLog() string {
	return _m.Called().String(0)
}

// Hint モック
func (_m *MockRummy500Interactor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockRummy500Interactor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
