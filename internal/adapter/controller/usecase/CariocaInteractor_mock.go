//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockCariocaInteractor モック
type MockCariocaInteractor struct {
	mock.Mock
}

func (_m *MockCariocaInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockCariocaInteractor) ResetWithConfig(cfg domain.CariocaConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockCariocaInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockCariocaInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockCariocaInteractor) MeldContract(indicesPerSlot [][]int) string {
	return _m.Called(indicesPerSlot).String(0)
}

func (_m *MockCariocaInteractor) MeldExtra(indices []int) string {
	return _m.Called(indices).String(0)
}

func (_m *MockCariocaInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	return _m.Called(targetPlayerIdx, meldIdx, cardIndex).String(0)
}

func (_m *MockCariocaInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockCariocaInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockCariocaInteractor) GetConfig() domain.CariocaConfig {
	return _m.Called().Get(0).(domain.CariocaConfig)
}

func (_m *MockCariocaInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockCariocaInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
