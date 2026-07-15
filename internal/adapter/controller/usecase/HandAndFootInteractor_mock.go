//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockHandAndFootInteractor モック
type MockHandAndFootInteractor struct {
	mock.Mock
}

func (_m *MockHandAndFootInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) ResetWithConfig(cfg domain.HandAndFootConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockHandAndFootInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockHandAndFootInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockHandAndFootInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockHandAndFootInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) GetConfig() domain.HandAndFootConfig {
	return _m.Called().Get(0).(domain.HandAndFootConfig)
}

func (_m *MockHandAndFootInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockHandAndFootInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockHandAndFootInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
