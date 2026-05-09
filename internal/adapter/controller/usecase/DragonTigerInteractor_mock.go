//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockDragonTigerInteractor ドラゴンタイガーインタラクターモック
type MockDragonTigerInteractor struct {
	mock.Mock
}

func (m *MockDragonTigerInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDragonTigerInteractor) Bet(amount, betType int) string {
	args := m.Called(amount, betType)
	return args.String(0)
}

func (m *MockDragonTigerInteractor) ClearHistory() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDragonTigerInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockDragonTigerInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
