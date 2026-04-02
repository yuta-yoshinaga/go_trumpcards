//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCanastaInteractor モック
type MockCanastaInteractor struct {
	mock.Mock
}

func (_m *MockCanastaInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCanastaInteractor) ResetWithConfig(cfg domain.CanastaConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockCanastaInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockCanastaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockCanastaInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockCanastaInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockCanastaInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockCanastaInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockCanastaInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockCanastaInteractor) GetConfig() domain.CanastaConfig {
	return _m.Called().Get(0).(domain.CanastaConfig)
}

func (_m *MockCanastaInteractor) ActionLog() string {
	return _m.Called().String(0)
}
