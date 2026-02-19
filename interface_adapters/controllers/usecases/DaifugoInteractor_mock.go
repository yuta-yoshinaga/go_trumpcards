package usecases

import "github.com/stretchr/testify/mock"

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
