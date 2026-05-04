//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCasinoWarInteractor カジノウォーインタラクターモック
type MockCasinoWarInteractor struct {
	mock.Mock
}

func (m *MockCasinoWarInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoWarInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockCasinoWarInteractor) Surrender() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoWarInteractor) War() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCasinoWarInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockCasinoWarInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
