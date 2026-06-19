//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockFaroInteractor はファロインタラクターのモック。
type MockFaroInteractor struct {
	mock.Mock
}

func (m *MockFaroInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFaroInteractor) NextRound() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFaroInteractor) PlaceBet(rank, amount int, copper bool) string {
	args := m.Called(rank, amount, copper)
	return args.String(0)
}

func (m *MockFaroInteractor) ClearBet(rank int) string {
	args := m.Called(rank)
	return args.String(0)
}

func (m *MockFaroInteractor) ClearAll() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFaroInteractor) DealTurn() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFaroInteractor) Call(order []int) string {
	args := m.Called(order)
	return args.String(0)
}

func (m *MockFaroInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック。
func (m *MockFaroInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
