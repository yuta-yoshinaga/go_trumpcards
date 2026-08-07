//go:build test

package usecase

import "github.com/stretchr/testify/mock"

// MockLetItRideInteractor レット・イット・ライドインタラクターモック
type MockLetItRideInteractor struct {
	mock.Mock
}

func (m *MockLetItRideInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLetItRideInteractor) Bet(amount int) string {
	args := m.Called(amount)
	return args.String(0)
}

func (m *MockLetItRideInteractor) Pull() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLetItRideInteractor) PullConfirm() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLetItRideInteractor) LetItRide() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLetItRideInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}

// Snapshot モック
func (m *MockLetItRideInteractor) Snapshot() ([]byte, error) {
	ret := m.Called()
	return ret.Get(0).([]byte), ret.Error(1)
}
