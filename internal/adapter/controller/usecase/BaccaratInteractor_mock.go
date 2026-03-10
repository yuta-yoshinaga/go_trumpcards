package usecase

import "github.com/stretchr/testify/mock"

// MockBaccaratInteractor バカラインタラクターモック
type MockBaccaratInteractor struct {
	mock.Mock
}

func (m *MockBaccaratInteractor) Reset() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBaccaratInteractor) Bet(amount, betType int) string {
	args := m.Called(amount, betType)
	return args.String(0)
}

func (m *MockBaccaratInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}
