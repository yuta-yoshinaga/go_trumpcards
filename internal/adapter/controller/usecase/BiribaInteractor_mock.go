//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBiribaInteractor モック
type MockBiribaInteractor struct {
	mock.Mock
}

func (_m *MockBiribaInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) ResetWithConfig(cfg domain.BiribaConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockBiribaInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockBiribaInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockBiribaInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockBiribaInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) GetConfig() domain.BiribaConfig {
	return _m.Called().Get(0).(domain.BiribaConfig)
}

func (_m *MockBiribaInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockBiribaInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockBiribaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
