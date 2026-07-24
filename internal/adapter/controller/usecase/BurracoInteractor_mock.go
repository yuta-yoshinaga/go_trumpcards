//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBurracoInteractor モック
type MockBurracoInteractor struct {
	mock.Mock
}

func (_m *MockBurracoInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) ResetWithConfig(cfg domain.BurracoConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockBurracoInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockBurracoInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockBurracoInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockBurracoInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) GetConfig() domain.BurracoConfig {
	return _m.Called().Get(0).(domain.BurracoConfig)
}

func (_m *MockBurracoInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockBurracoInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockBurracoInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
