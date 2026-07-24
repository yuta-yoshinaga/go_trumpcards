//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockHighCardFlushInteractor ハイカードフラッシュインタラクターモック
type MockHighCardFlushInteractor struct {
	mock.Mock
}

func (m *MockHighCardFlushInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockHighCardFlushInteractor) Bet(ante, flushBonus, straightFlush int) string {
	args := m.Called(ante, flushBonus, straightFlush)
	return args.String(0)
}

func (m *MockHighCardFlushInteractor) Raise(multiplier int) string {
	args := m.Called(multiplier)
	return args.String(0)
}

func (m *MockHighCardFlushInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockHighCardFlushInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Hint モック
func (m *MockHighCardFlushInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockHighCardFlushInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
