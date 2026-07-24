//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockVideoPokerInteractor ビデオポーカーインタラクターモック
type MockVideoPokerInteractor struct {
	mock.Mock
}

func (m *MockVideoPokerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVideoPokerInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockVideoPokerInteractor) Hold(indices []int) string {
	args := m.Called(indices)
	return args.String(0)
}

func (m *MockVideoPokerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Hint モック
func (m *MockVideoPokerInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockVideoPokerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
