//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockUltimateTexasHoldemInteractor アルティメット・テキサスホールデムインタラクターモック
type MockUltimateTexasHoldemInteractor struct {
	mock.Mock
}

func (m *MockUltimateTexasHoldemInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) Bet(ante, trips int) string {
	args := m.Called(ante, trips)
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) Play(multiplier int) string {
	args := m.Called(multiplier)
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) Check() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockUltimateTexasHoldemInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockUltimateTexasHoldemInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
