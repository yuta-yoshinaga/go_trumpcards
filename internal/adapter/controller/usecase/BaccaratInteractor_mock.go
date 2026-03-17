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

func (m *MockBaccaratInteractor) Bet(amount, betType, ppBet, bpBet int) string {
	args := m.Called(amount, betType, ppBet, bpBet)
	return args.String(0)
}

func (m *MockBaccaratInteractor) ClearHistory() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockBaccaratInteractor) ActionLog() string {
	args := m.Called()
	return args.String(0)
}
