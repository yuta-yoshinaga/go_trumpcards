//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockOasisPokerInteractor オアシスポーカーインタラクターモック
type MockOasisPokerInteractor struct {
	mock.Mock
}

func (m *MockOasisPokerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOasisPokerInteractor) Bet(ante, jackpot int) string {
	args := m.Called(ante, jackpot)
	return args.String(0)
}

func (m *MockOasisPokerInteractor) Exchange(indices []int) string {
	args := m.Called(indices)
	return args.String(0)
}

func (m *MockOasisPokerInteractor) Stand() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOasisPokerInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOasisPokerInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOasisPokerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Hint モック
func (m *MockOasisPokerInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockOasisPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
