//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockPaiGowInteractor パイガオポーカーインタラクターモック
type MockPaiGowInteractor struct {
	mock.Mock
}

func (m *MockPaiGowInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPaiGowInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockPaiGowInteractor) SetHands(lowIdx0, lowIdx1 int) string {
	args := m.Called(lowIdx0, lowIdx1)
	return args.String(0)
}

func (m *MockPaiGowInteractor) AutoSetHands() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPaiGowInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPaiGowInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockPaiGowInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
