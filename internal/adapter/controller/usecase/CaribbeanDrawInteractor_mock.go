//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockCaribbeanDrawInteractor カリビアン・ドロー・ポーカーインタラクターモック
type MockCaribbeanDrawInteractor struct {
	mock.Mock
}

func (m *MockCaribbeanDrawInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) Bet(ante, jackpot int) string {
	args := m.Called(ante, jackpot)
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) Draw(indices []int) string {
	args := m.Called(indices)
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) Play() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) Fold() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCaribbeanDrawInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockCaribbeanDrawInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
