//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSambaInteractor モック
type MockSambaInteractor struct {
	mock.Mock
}

func (_m *MockSambaInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockSambaInteractor) ResetWithConfig(cfg domain.SambaConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockSambaInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockSambaInteractor) DrawFromDiscard(naturalPairIndices []int) string {
	return _m.Called(naturalPairIndices).String(0)
}

func (_m *MockSambaInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockSambaInteractor) SkipMeld() string {
	return _m.Called().String(0)
}

func (_m *MockSambaInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockSambaInteractor) GoOut() string {
	return _m.Called().String(0)
}

func (_m *MockSambaInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockSambaInteractor) GetConfig() domain.SambaConfig {
	return _m.Called().Get(0).(domain.SambaConfig)
}

func (_m *MockSambaInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockSambaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
