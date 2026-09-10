//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockBassetInteractor はバセットインタラクターのモック。
type MockBassetInteractor struct {
	mock.Mock
}

func (m *MockBassetInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBassetInteractor) NextRound() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBassetInteractor) PlaceBet(rank, amount int) string {
	args := m.Called(rank, amount)
	return args.String(0)
}

func (m *MockBassetInteractor) DealTurn() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBassetInteractor) TakeWinnings() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBassetInteractor) PressParoli() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBassetInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック。
func (m *MockBassetInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
