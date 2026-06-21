//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockKalookiInteractor モック
type MockKalookiInteractor struct {
	mock.Mock
}

func (_m *MockKalookiInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockKalookiInteractor) ResetWithConfig(cfg domain.KalookiConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockKalookiInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockKalookiInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockKalookiInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockKalookiInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	return _m.Called(targetPlayerIdx, meldIdx, cardIndex).String(0)
}

func (_m *MockKalookiInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockKalookiInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockKalookiInteractor) GetConfig() domain.KalookiConfig {
	return _m.Called().Get(0).(domain.KalookiConfig)
}

func (_m *MockKalookiInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockKalookiInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
