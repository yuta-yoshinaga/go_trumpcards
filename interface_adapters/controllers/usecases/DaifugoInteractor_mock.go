package usecases

import (
	"github.com/stretchr/testify/mock"
)

// MockDaifugoInteractor 大富豪インタラクターモック
type MockDaifugoInteractor struct {
	mock.Mock
}

// Reset モック実装
func (m *MockDaifugoInteractor) Reset() interface{} {
	args := m.Called()
	return args.Get(0)
}

// Play モック実装
func (m *MockDaifugoInteractor) Play(cardIndices []int) interface{} {
	args := m.Called(cardIndices)
	return args.Get(0)
}

// Pass モック実装
func (m *MockDaifugoInteractor) Pass() interface{} {
	args := m.Called()
	return args.Get(0)
}
