package usecases

import (
	"github.com/stretchr/testify/mock"
)

// MockDaifugoInteractor 大富豪インタラクターモック
type MockDaifugoInteractor struct {
	mock.Mock
}

// Reset モック実装
func (m *MockDaifugoInteractor) Reset() string {
	args := m.Called()
	return args.Get(0).(string)
}

// Play モック実装
func (m *MockDaifugoInteractor) Play(cardIndices []int) string {
	args := m.Called(cardIndices)
	return args.Get(0).(string)
}

// Pass モック実装
func (m *MockDaifugoInteractor) Pass() string {
	args := m.Called()
	return args.Get(0).(string)
}
