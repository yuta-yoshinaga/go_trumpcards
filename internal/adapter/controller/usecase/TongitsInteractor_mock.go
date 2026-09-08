//go:build test
// +build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockTongitsInteractor モック
type MockTongitsInteractor struct {
	mock.Mock
}

func (_m *MockTongitsInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockTongitsInteractor) ResetWithConfig(cfg domain.TongitsConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockTongitsInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockTongitsInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockTongitsInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockTongitsInteractor) Meld(indices []int) string {
	return _m.Called(indices).String(0)
}

func (_m *MockTongitsInteractor) Sapaw(targetPlayerIdx, meldIdx, cardIndex int) string {
	return _m.Called(targetPlayerIdx, meldIdx, cardIndex).String(0)
}

func (_m *MockTongitsInteractor) Challenge(agreed []bool) string {
	return _m.Called(agreed).String(0)
}

func (_m *MockTongitsInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockTongitsInteractor) GetConfig() domain.TongitsConfig {
	return _m.Called().Get(0).(domain.TongitsConfig)
}

func (_m *MockTongitsInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockTongitsInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
