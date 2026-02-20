package usecases

import "github.com/stretchr/testify/mock"

// MockSevensInteractor 7並べインタラクターモック
type MockSevensInteractor struct {
	mock.Mock
}

// Reset モック
func (_m *MockSevensInteractor) Reset() string {
	ret := _m.Called()
	return ret.Get(0).(string)
}

// Play モック
func (_m *MockSevensInteractor) Play(idx int) string {
	ret := _m.Called(idx)
	return ret.Get(0).(string)
}
