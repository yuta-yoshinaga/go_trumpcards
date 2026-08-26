//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockBoliviaInteractor モック
type MockBoliviaInteractor struct {
	mock.Mock
}

func (_m *MockBoliviaInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockBoliviaInteractor) ResetWithConfig(cfg domain.BoliviaConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockBoliviaInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockBoliviaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockBoliviaInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockBoliviaInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockBoliviaInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockBoliviaInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockBoliviaInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockBoliviaInteractor) GetConfig() domain.BoliviaConfig {
	return _m.Called().Get(0).(domain.BoliviaConfig)
}

func (_m *MockBoliviaInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockBoliviaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
