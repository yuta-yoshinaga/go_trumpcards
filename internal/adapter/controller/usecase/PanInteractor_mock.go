//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockPanInteractor モック
type MockPanInteractor struct {
	mock.Mock
}

func (_m *MockPanInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockPanInteractor) ResetWithConfig(cfg domain.PanConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockPanInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockPanInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockPanInteractor) Meld(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockPanInteractor) Layoff(meldOwner, meldIdx, cardIndex int) string {
	return _m.Called(meldOwner, meldIdx, cardIndex).String(0)
}

func (_m *MockPanInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockPanInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockPanInteractor) GetConfig() domain.PanConfig {
	return _m.Called().Get(0).(domain.PanConfig)
}

func (_m *MockPanInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockPanInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
