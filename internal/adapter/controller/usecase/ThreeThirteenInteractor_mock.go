//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockThreeThirteenInteractor モック
type MockThreeThirteenInteractor struct {
	mock.Mock
}

func (_m *MockThreeThirteenInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockThreeThirteenInteractor) ResetWithConfig(cfg domain.ThreeThirteenConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockThreeThirteenInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockThreeThirteenInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockThreeThirteenInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockThreeThirteenInteractor) Knock(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockThreeThirteenInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockThreeThirteenInteractor) GetConfig() domain.ThreeThirteenConfig {
	return _m.Called().Get(0).(domain.ThreeThirteenConfig)
}

func (_m *MockThreeThirteenInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockThreeThirteenInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
