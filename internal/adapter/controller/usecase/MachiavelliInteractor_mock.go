//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockMachiavelliInteractor モック
type MockMachiavelliInteractor struct {
	mock.Mock
}

func (_m *MockMachiavelliInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockMachiavelliInteractor) ResetWithConfig(cfg domain.MachiavelliConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockMachiavelliInteractor) Draw() string {
	return _m.Called().String(0)
}

func (_m *MockMachiavelliInteractor) Play(refs [][]domain.MachiavelliCardRef, handIndices []int) string {
	return _m.Called(refs, handIndices).String(0)
}

func (_m *MockMachiavelliInteractor) NewMeld(handIndices []int) string {
	return _m.Called(handIndices).String(0)
}

func (_m *MockMachiavelliInteractor) Layoff(meldIdx, handIndex int) string {
	return _m.Called(meldIdx, handIndex).String(0)
}

func (_m *MockMachiavelliInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockMachiavelliInteractor) GetConfig() domain.MachiavelliConfig {
	return _m.Called().Get(0).(domain.MachiavelliConfig)
}

func (_m *MockMachiavelliInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockMachiavelliInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
