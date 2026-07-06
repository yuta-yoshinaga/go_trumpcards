//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockOichoKabuInteractor おいちょかぶインタラクターモック
type MockOichoKabuInteractor struct {
	mock.Mock
}

func (m *MockOichoKabuInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOichoKabuInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockOichoKabuInteractor) Draw() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOichoKabuInteractor) Stand() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOichoKabuInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockOichoKabuInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
