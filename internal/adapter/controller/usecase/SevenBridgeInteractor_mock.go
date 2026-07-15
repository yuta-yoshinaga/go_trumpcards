//go:build test

package usecase

import (
	"github.com/stretchr/testify/mock"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockSevenBridgeInteractor モック
type MockSevenBridgeInteractor struct {
	mock.Mock
}

func (_m *MockSevenBridgeInteractor) Reset() string {
	return _m.Called().String(0)
}

func (_m *MockSevenBridgeInteractor) ResetWithConfig(cfg domain.SevenBridgeConfig) string {
	return _m.Called(cfg).String(0)
}

func (_m *MockSevenBridgeInteractor) DrawFromStock() string {
	return _m.Called().String(0)
}

func (_m *MockSevenBridgeInteractor) ClaimPon(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockSevenBridgeInteractor) ClaimChi(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockSevenBridgeInteractor) Meld(cardIndices []int) string {
	return _m.Called(cardIndices).String(0)
}

func (_m *MockSevenBridgeInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	return _m.Called(targetPlayerIdx, meldIdx, cardIndex).String(0)
}

func (_m *MockSevenBridgeInteractor) Discard(cardIndex int) string {
	return _m.Called(cardIndex).String(0)
}

func (_m *MockSevenBridgeInteractor) NextRound() string {
	return _m.Called().String(0)
}

func (_m *MockSevenBridgeInteractor) Hint() string {
	return _m.Called().String(0)
}

func (_m *MockSevenBridgeInteractor) GetConfig() domain.SevenBridgeConfig {
	return _m.Called().Get(0).(domain.SevenBridgeConfig)
}

func (_m *MockSevenBridgeInteractor) ActionLog() string {
	return _m.Called().String(0)
}

// Snapshot モック
func (_m *MockSevenBridgeInteractor) Snapshot() ([]byte, error) {
	ret := _m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
