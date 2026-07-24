//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCasinoHoldemInteractor カジノホールデムインタラクターモック
type MockCasinoHoldemInteractor struct {
	mock.Mock
}

func (m *MockCasinoHoldemInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoHoldemInteractor) Bet(ante, bonus int) string {
	args := m.Called(ante, bonus)
	return args.String(0)
}

func (m *MockCasinoHoldemInteractor) Call() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoHoldemInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoHoldemInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Hint モック
func (m *MockCasinoHoldemInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockCasinoHoldemInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
