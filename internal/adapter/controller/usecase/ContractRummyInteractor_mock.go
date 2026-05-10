//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockContractRummyInteractor モック
type MockContractRummyInteractor struct {
	mock.Mock
}

func (_m *MockContractRummyInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockContractRummyInteractor) ResetWithConfig(cfg domain.ContractRummyConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockContractRummyInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockContractRummyInteractor) DrawFromDiscard() string {
	return _m.Called().String(0)
}

func (_m *MockContractRummyInteractor) MeldContract(indicesPerSlot [][]int) string {
	return _m.Called(indicesPerSlot).String(0)
}

func (_m *MockContractRummyInteractor) MeldExtra(indices []int) string {
	return _m.Called(indices).String(0)
}

func (_m *MockContractRummyInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	return _m.Called(targetPlayerIdx, meldIdx, cardIndex).String(0)
}

func (_m *MockContractRummyInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockContractRummyInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockContractRummyInteractor) GetConfig() domain.ContractRummyConfig {
	return _m.Called().Get(0).(domain.ContractRummyConfig)
}

func (_m *MockContractRummyInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockContractRummyInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
