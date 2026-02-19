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
	return args.String(0)
}

// Play モック実装
func (m *MockDaifugoInteractor) Play(cardIndices []int) string {
	args := m.Called(cardIndices)
	return args.String(0)
}

// Pass モック実装
func (m *MockDaifugoInteractor) Pass() string {
	args := m.Called()
	return args.String(0)
}
