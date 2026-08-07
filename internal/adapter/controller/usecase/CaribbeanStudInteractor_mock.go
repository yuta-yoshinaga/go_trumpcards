//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCaribbeanStudInteractor カリビアンスタッドポーカーインタラクターモック
type MockCaribbeanStudInteractor struct {
	mock.Mock
}

func (m *MockCaribbeanStudInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanStudInteractor) Bet(ante, jackpot int) string {
	args := m.Called(ante, jackpot)
	return args.String(0)
}

func (m *MockCaribbeanStudInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanStudInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanStudInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanStudInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockCaribbeanStudInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
