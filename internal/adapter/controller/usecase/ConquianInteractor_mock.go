//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockConquianInteractor モック
type MockConquianInteractor struct {
	mock.Mock
}

func (_m *MockConquianInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockConquianInteractor) ResetWithConfig(cfg domain.ConquianConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockConquianInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockConquianInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockConquianInteractor) Meld(meldGroups [][]int) string {
	return _m.Called(meldGroups).String(0)
}

func (_m *MockConquianInteractor) MeldWithTargets(meldGroups [][]int, extendTargets []int) string {
	return _m.Called(meldGroups, extendTargets).String(0)
}

func (_m *MockConquianInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockConquianInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockConquianInteractor) GetConfig() domain.ConquianConfig {
	return _m.Called().Get(0).(domain.ConquianConfig)
}

func (_m *MockConquianInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockConquianInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
