//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFourCardPokerInteractor is the testify mock for FourCardPokerInteractorIF.
type MockFourCardPokerInteractor struct {
	mock.Mock
}

func (m *MockFourCardPokerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFourCardPokerInteractor) Bet(ante, acesUp int) string {
	args := m.Called(ante, acesUp)
	return args.String(0)
}

func (m *MockFourCardPokerInteractor) Play(multiplier int) string {
	args := m.Called(multiplier)
	return args.String(0)
}

func (m *MockFourCardPokerInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFourCardPokerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFourCardPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
