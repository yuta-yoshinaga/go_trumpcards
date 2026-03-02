package usecase

import (
	"github.com/stretchr/testify/mock"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// MockDaifugoInteractor 大富豪インタラクターモック
type MockDaifugoInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockDaifugoInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockDaifugoInteractor) Play(indices []int) string {
	ret := _m.Called(indices)
	return ret.Get(0).(string)
}

// ResetWithConfig モック
func (_m *MockDaifugoInteractor) ResetWithConfig(config domain.DaifugoConfig) string {
	ret := _m.Called(config)
	return ret.Get(0).(string)
}

// Sort モック
func (_m *MockDaifugoInteractor) Sort(mode domain.DaifugoSortMode) string {
	ret := _m.Called(mode)
	return ret.Get(0).(string)
}
