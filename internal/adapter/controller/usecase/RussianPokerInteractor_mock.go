//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockRussianPokerInteractor ロシアンポーカーインタラクターモック
type MockRussianPokerInteractor struct {
	mock.Mock
}

func (m *MockRussianPokerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Bet(ante int) string {
	args := m.Called(ante)
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Exchange(indices []int) string {
	args := m.Called(indices)
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Buy6th() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Select(discardIndex int) string {
	args := m.Called(discardIndex)
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) ForceExchange() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) Decline() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRussianPokerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockRussianPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
