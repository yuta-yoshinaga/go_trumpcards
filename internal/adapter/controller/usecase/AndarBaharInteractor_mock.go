//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockAndarBaharInteractor アンダーバハールインタラクターモック
type MockAndarBaharInteractor struct {
	mock.Mock
}

func (m *MockAndarBaharInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAndarBaharInteractor) Bet(amount, target, sideAmount, sideBand int) string {
	args := m.Called(amount, target, sideAmount, sideBand)
	return args.String(0)
}

func (m *MockAndarBaharInteractor) ClearHistory() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAndarBaharInteractor) Hint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAndarBaharInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAndarBaharInteractor) Snapshot() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
