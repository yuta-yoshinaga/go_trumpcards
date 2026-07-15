//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockRedDogInteractor レッドドッグインタラクターモック
type MockRedDogInteractor struct {
	mock.Mock
}

func (m *MockRedDogInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRedDogInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockRedDogInteractor) Raise(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockRedDogInteractor) Stay() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRedDogInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRedDogInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockRedDogInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
