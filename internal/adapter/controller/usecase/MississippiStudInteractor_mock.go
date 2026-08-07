//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockMississippiStudInteractor ミシシッピ・スタッドインタラクターモック
type MockMississippiStudInteractor struct {
	mock.Mock
}

func (m *MockMississippiStudInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMississippiStudInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockMississippiStudInteractor) Play(multiplier int) string {
	args := m.Called(multiplier)
	return args.String(0)
}

func (m *MockMississippiStudInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMississippiStudInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockMississippiStudInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockMississippiStudInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
